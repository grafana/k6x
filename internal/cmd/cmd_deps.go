// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/szkiba/k6x/internal/dependency"
	"github.com/szkiba/k6x/internal/resolver"
)

func depsCommand(
	ctx context.Context,
	res resolver.Resolver,
	opts *options,
	out *os.File, //nolint:forbidigo
) error {
	if opts.help {
		return usage(out, depsUsage, opts)
	}

	var result interface{}

	var deps dependency.Dependencies
	var err error

	deps, err = collectDependencies(ctx, res, opts)
	if err != nil {
		return err
	}

	if opts.resolve {
		ings, rerr := res.Resolve(ctx, deps)
		if rerr != nil {
			return rerr
		}

		result = ings
	} else {
		result = deps
	}

	if opts.json {
		encoder := json.NewEncoder(out)
		encoder.SetEscapeHTML(false)

		return encoder.Encode(result)
	}

	_, err = fmt.Fprint(out, result)
	if err != nil {
		return err
	}

	return nil
}

const depsUsage = `Print k6 and extension dependencies for a script.

Usage:
  {{.appname}} deps [flags] [script]

Flags:
  -o, --out name     output extension name
  --json             use JSON output format
  --resolve          print resolved dependencies
  --with dependency  additional dependency and version constraints

  -h, --help      display this help
`
