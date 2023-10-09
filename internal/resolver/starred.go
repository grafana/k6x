// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

package resolver

import (
	"context"
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/google/go-github/v55/github"
	"github.com/sirupsen/logrus"
	"github.com/szkiba/k6x/internal/dependency"
)

func (res *ghResolver) getStarred(ctx context.Context, stars int) (map[string]struct{}, error) {
	result, _, err := res.client.Search.Repositories(
		ctx,
		"topic:xk6",
		&github.SearchOptions{ListOptions: github.ListOptions{PerPage: 100}},
	)
	if err != nil {
		return nil, err
	}

	match := make(map[string]struct{})

	for _, repo := range result.Repositories {
		if repo.GetStargazersCount() <= stars || repo.GetArchived() {
			continue
		}

		match["github.com/"+repo.GetFullName()] = struct{}{}
	}

	return match, nil
}

func (res *ghResolver) Starred(ctx context.Context, stars int) (dependency.Modules, error) {
	if len(getToken()) == 0 {
		return nil, fmt.Errorf("%w: GitHub authentication required", ErrResolver)
	}

	logrus.Info("filtering extensions")

	reg, err := res.getRegistry(ctx)
	if err != nil {
		return nil, err
	}

	starred, err := res.getStarred(ctx, stars)
	if err != nil {
		return nil, err
	}

	candidates := make(dependency.Modules)

	for _, mod := range reg.toUniqueModules() {
		if _, ok := starred[mod.Path]; ok {
			candidates[mod.Name] = mod
		}
	}
	candidates[k6] = &dependency.Module{Artifact: &dependency.Artifact{Name: k6}}

	found := make(dependency.Modules)
	constraint, _ := semver.NewConstraint("*")

	for _, mod := range candidates {
		logrus.Infof("resolving latest version for %s", mod.Name)

		var owner, repo string

		if mod.Name == "k6" {
			owner = "grafana"
			repo = "k6"
		} else {
			parts := strings.SplitN(mod.Path, "/", 4)

			owner = parts[1]
			repo = parts[2]
		}

		tags, _, err := res.client.Repositories.ListTags(
			ctx,
			owner,
			repo,
			&github.ListOptions{PerPage: 100},
		)
		if err != nil {
			return nil, err
		}

		if len(tags) == 0 {
			continue
		}

		for _, tag := range tags {
			name := tag.GetName()
			if name[0] != 'v' {
				continue
			}

			ver, err := semver.NewVersion(name)
			if err != nil {
				continue
			}

			if constraint.Check(ver) {
				mod.Version = ver
				break
			}
		}

		if mod.Version != nil {
			found[mod.Name] = mod
		}
	}

	return found, nil
}
