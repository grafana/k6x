// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/szkiba/k6x/internal/dependency"
	"github.com/szkiba/k6x/internal/resolver"
)

//nolint:gochecknoglobals
var (
	_appname = "k6x"
	_version = "devel"
)

func versionCommand(
	ctx context.Context,
	cmd string,
	res resolver.Resolver,
	opts *options,
	stdin, stdout, stderr *os.File, //nolint:forbidigo
) (int, error) {
	if err := prepare(ctx, cmd, make(dependency.Dependencies), res, opts); err != nil {
		return exitErr, err
	}

	opts.spinner.Stop()

	if opts.help {
		_ = usage(stdout, versionUsage, opts)
	} else {
		fmt.Fprintf(stdout, "%s %s (%s)\n", _appname, _version, cmd)
	}

	opts.argv[0] = cmd

	return opts.exec(cmd, opts.argv, stdin, stdout, stderr)
}

//nolint:gochecknoglobals
var versionUsage = `Launcher Flags:
  --bin-dir path  cache folder for k6 binary (default: {{.bin}})
  --builder list  comma separated list of builders (default: {{.builders}})
  --clean         remove cached k6 binary
  --dry           do not run k6 command

`
