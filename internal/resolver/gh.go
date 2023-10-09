// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

//nolint:revive
package resolver

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/google/go-github/v55/github"
	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
	"github.com/jmespath/go-jmespath"
	"github.com/szkiba/k6x/internal/dependency"
)

type ghResolver struct {
	client *github.Client
	filter *jmespath.JMESPath
}

func New(cachedir string, filter string) (Resolver, error) {
	transport := httpcache.NewTransport(diskcache.New(cachedir))

	client := &http.Client{Transport: newTransport(transport)}

	res := new(ghResolver)

	res.client = github.NewClient(client)

	if len(filter) != 0 {
		query, err := jmespath.Compile(filter)
		if err != nil {
			return nil, err
		}

		res.filter = query
	}

	return res, nil
}

func (res *ghResolver) Resolve(
	ctx context.Context,
	deps dependency.Dependencies,
) (dependency.Modules, error) {
	mods, err := res.resolveModules(ctx, deps)
	if err != nil {
		// partially resolved idgredients returned also with error
		return mods, err
	}

	if err := res.resolveReleases(ctx, deps, mods); err != nil {
		// partially resolved idgredients returned also with error
		return mods, err
	}

	return mods, nil
}

func (res *ghResolver) getRegistry(ctx context.Context) (*extensionRegistry, error) {
	content, _, _, err := res.client.Repositories.GetContents(
		ctx,
		"grafana",
		"k6-docs",
		"src/data/doc-extensions/extensions.json",
		nil,
	)
	if err != nil {
		return nil, err
	}

	str, err := content.GetContent()
	if err != nil {
		return nil, err
	}

	reg, err := parseExtensionRegistry([]byte(str), res.filter)
	if err != nil {
		return nil, err
	}

	return reg, nil
}

func (res *ghResolver) resolveModules(
	ctx context.Context,
	deps dependency.Dependencies,
) (dependency.Modules, error) {
	reg, err := res.getRegistry(ctx)
	if err != nil {
		return nil, err
	}

	mods := make(dependency.Modules)

	mods[k6] = &dependency.Module{Artifact: &dependency.Artifact{Name: k6}}

	for name, mod := range reg.toModules() {
		if _, ok := deps[name]; ok {
			mods[name] = mod
		}
	}

	err = checkForMisingPaths(deps, mods)

	// partially resolved idgredients returned also with error
	return mods, err
}

func checkForMisingPaths(deps dependency.Dependencies, ings dependency.Modules) error {
	missing := make(dependency.Dependencies)

	for _, dep := range deps {
		mod, ok := ings[dep.Name]
		if !ok || (mod.Name != k6 && len(mod.Path) == 0) {
			missing[dep.Name] = dep
		}
	}

	if len(missing) == 0 {
		return nil
	}

	return fmt.Errorf("%w: unable to resolve module: %s", ErrResolver, missing)
}

func checkForMisingVersions(deps dependency.Dependencies, mods dependency.Modules) error {
	missing := make(dependency.Dependencies)

	for _, dep := range deps {
		mod, ok := mods[dep.Name]
		if !ok || !dep.Check(mod.Version) {
			missing[dep.Name] = dep
		}
	}

	if len(missing) == 0 {
		return nil
	}

	return fmt.Errorf("%w: unable to fulfill constraints: %s", ErrResolver, missing)
}

type ghTransport struct {
	base  http.RoundTripper
	token string
}

func newTransport(base http.RoundTripper) *ghTransport {
	return &ghTransport{base: base, token: getToken()}
}

//nolint:gosec,forbidigo
func getToken() string {
	if token := os.Getenv(envAppToken); len(token) != 0 {
		return token
	}

	if token := os.Getenv(envGhToken); len(token) != 0 {
		return token
	}

	if token := os.Getenv(envGitHubToken); len(token) != 0 {
		return token
	}

	gh := os.Getenv(envAppGhPath)
	if gh == "" {
		gh = os.Getenv(envGhPath)
	}

	if gh == "" {
		gh, _ = exec.LookPath(ghExe)
	}

	if len(gh) == 0 {
		return ""
	}

	result, err := exec.Command(gh, "auth", "token", "--secure-storage", "--hostname", ghHost).
		Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(result))
}

func (t *ghTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if len(t.token) != 0 && len(req.Header.Get(hdrAuthorization)) == 0 {
		req.Header.Set(hdrAuthorization, "token "+t.token)
	}

	return t.base.RoundTrip(req)
}

const (
	ghHost         = "github.com"
	ghExe          = "gh"
	envAppToken    = "K6X_GITHUB_TOKEN" //nolint:gosec
	envGhToken     = "GH_TOKEN"
	envGitHubToken = "GITHUB_TOKEN" //nolint:gosec

	envAppGhPath = "K6X_GH_PATH"
	envGhPath    = "GH_PATH"

	hdrAuthorization = "Authorization"
)
