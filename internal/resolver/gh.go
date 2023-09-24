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
) (Ingredients, error) {
	ings, err := res.resolveModules(ctx, deps)
	if err != nil {
		// partially resolved idgredients returned also with error
		return ings, err
	}

	if err := res.resolveReleases(ctx, deps, ings); err != nil {
		// partially resolved idgredients returned also with error
		return ings, err
	}

	return ings, nil
}

func (res *ghResolver) resolveModules(
	ctx context.Context,
	deps dependency.Dependencies,
) (Ingredients, error) {
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

	ings := make(Ingredients)

	ings[k6] = k6DefaultIngredient

	for name, ing := range reg.toIngredients() {
		if _, ok := deps[name]; ok {
			ings[name] = ing
		}
	}

	err = checkForMisingModules(deps, ings)

	// partially resolved idgredients returned also with error
	return ings, err
}

func checkForMisingModules(deps dependency.Dependencies, ings Ingredients) error {
	missing := make(dependency.Dependencies)

	for _, dep := range deps {
		ing, ok := ings[dep.Name]
		if !ok || (ing.Name != k6 && len(ing.Module) == 0) {
			missing[dep.Name] = dep
		}
	}

	if len(missing) == 0 {
		return nil
	}

	return fmt.Errorf("%w: unable to resolve module: %s", ErrResolver, missing)
}

func checkForMisingVersions(deps dependency.Dependencies, ings Ingredients) error {
	missing := make(dependency.Dependencies)

	for _, dep := range deps {
		ing, ok := ings[dep.Name]
		if !ok || !dep.Check(ing.Version) {
			missing[dep.Name] = dep
		}
	}

	if len(missing) == 0 {
		return nil
	}

	return fmt.Errorf("%w: unable to fulfill constraints: %s", ErrResolver, missing)
}
