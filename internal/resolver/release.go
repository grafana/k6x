// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

//nolint:revive
package resolver

import (
	"context"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/google/go-github/v55/github"
	"github.com/szkiba/k6x/internal/dependency"
)

func (res *ghResolver) resolveReleases(
	ctx context.Context,
	deps dependency.Dependencies,
	mods dependency.Modules,
) error {
	for _, mod := range mods {
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
			return err
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

			dep, ok := deps[mod.Name]

			if ok && dep.Check(ver) {
				mod.Version = ver
				break
			}
		}
	}

	if err := checkForMisingVersions(deps, mods); err != nil {
		return err
	}

	return nil
}
