// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

package cmd

import (
	"context"
	"io"

	"github.com/szkiba/k6x/internal/builder"
	"github.com/szkiba/k6x/internal/dependency"
	"github.com/szkiba/k6x/internal/resolver"
)

func preloadCommand(
	ctx context.Context,
	res resolver.Resolver,
	opts *options,
	out io.Writer,
) error {
	if opts.help {
		return usage(out, preloadUsage, opts)
	}

	var mods dependency.Modules
	var err error

	if len(opts.with) == 0 {
		mods, err = res.Starred(ctx, opts.stars)
	} else {
		ensureK6(opts.with)

		mods, err = res.Resolve(ctx, opts.with)
	}

	if err != nil {
		return err
	}

	b, err := builder.New(ctx, opts.engines...)
	if err != nil {
		return err
	}

	return builder.Preload(ctx, b, mods, opts.platforms)
}

const preloadUsage = `Preload the build cache with popular extensions.

Usage:
  {{.appname}} preload [flags]

Flags:
  --platform list    comma separated list of platforms (default: {{.platforms}})
  --stars number     minimum number of repository stargazers (default: 5)
  --with dependency  dependency and version constraints (default: latest version of k6 and registered extensions)
  --filter expr      jmespath syntax extension registry filter (default: [*])
  --builder list     comma separated list of builders (default: {{.builders}})
  -h, --help         display this help
`
