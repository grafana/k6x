// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

//nolint:revive
package resolver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/go-github/v55/github"
	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
	"github.com/szkiba/k6x/internal/dependency"
)

type ghResolver struct {
	client *github.Client
}

func NewWithCacheDir(cachedir string) Resolver {
	transport := httpcache.NewTransport(diskcache.New(cachedir))

	return NewWithHTTPClient(transport.Client())
}

func NewWithHTTPClient(client *http.Client) Resolver {
	return NewWithGitHubClient(github.NewClient(client))
}

func NewWithGitHubClient(client *github.Client) Resolver {
	return &ghResolver{client: client}
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

func (res *ghResolver) resolveModules(
	ctx context.Context,
	deps dependency.Dependencies,
) (dependency.Modules, error) {
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

	reg := new(extensionRegistry)

	if err = json.Unmarshal([]byte(str), reg); err != nil {
		return nil, err
	}

	mods := make(dependency.Modules)

	mods[k6] = &dependency.Module{Name: k6}

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
