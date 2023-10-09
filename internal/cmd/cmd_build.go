// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

package cmd

import (
	"context"
	"io"

	"github.com/szkiba/k6x/internal/dependency"
	"github.com/szkiba/k6x/internal/resolver"
)

func buildCommand(
	ctx context.Context,
	res resolver.Resolver,
	opts *options,
	out io.Writer,
) error {
	if opts.help {
		return usage(out, buildUsage, opts)
	}

	var deps dependency.Dependencies
	var err error

	if deps, err = collectDependencies(ctx, res, opts); err != nil {
		return err
	}

	return prepare(ctx, "", deps, res, opts)
}

const buildUsage = `Build custom k6 binary for a script.

Usage:
  {{.appname}} build [flags] [script]

Flags:
  -o, --out name     output extension name
  --bin-dir path     folder for custom k6 binary (default: {{.bin}})
  --with dependency  additional dependency and version constraints
  --builder list     comma separated list of builders (default: {{.builders}})
  --no-color         disable colored output  
  -h, --help         display this help
`
