// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

// Package cmd contains the command line tool main parts.
//
//nolint:forbidigo
package cmd

import (
	"context"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"text/template"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/szkiba/k6x/internal/resolver"
)

const exitErr = 116

// Main is the main entry point.
func Main(ctx context.Context, args []string, stdin, stdout, stderr *os.File, afs afero.Fs) int {
	code, err := main(ctx, args, stdin, stdout, stderr, afs)
	if err != nil {
		logrus.Error(err)
	}

	return code
}

func main(
	ctx context.Context,
	args []string,
	stdin, stdout, stderr *os.File,
	afs afero.Fs,
) (int, error) {
	opts, err := getopts(args, afs)
	if err != nil {
		return exitErr, err
	}

	defer opts.spinner.Stop()

	initLogger(opts)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			opts.spinner.Stop()
			os.Exit(1)
		}
	}()

	res, err := resolver.New(opts.dirs.http, opts.filter)
	if err != nil {
		return exitErr, err
	}

	cmd := filepath.Join(opts.dirs.bin, k6Binary)

	if opts.deps() {
		err = depsCommand(ctx, res, opts, stdout)
		if err == nil {
			return 0, nil
		}

		return exitErr, err
	}

	if opts.build() {
		err = buildCommand(ctx, res, opts, stdout)
		if err == nil {
			return 0, nil
		}

		return exitErr, err
	}

	if opts.service() {
		return 0, serviceCommand(ctx, res, opts, stdout)
	}

	if opts.preload() {
		return 0, preloadCommand(ctx, res, opts, stdout)
	}

	if opts.version() {
		return versionCommand(ctx, cmd, res, opts, stdin, stdout, stderr)
	}

	if opts.help || len(opts.argv) == 1 {
		usagelogo(stdout)

		if err := usage(stdout, otherUsage, opts); err != nil {
			return exitErr, err
		}
	}

	if opts.run() {
		return runCommand(ctx, cmd, res, opts, stdin, stdout, stderr)
	}

	return otherCommand(ctx, cmd, res, opts, stdin, stdout, stderr)
}

func usage(out io.Writer, tmpl string, opts *options) error {
	name := "usage"
	if len(opts.args) > 1 {
		name += opts.args[1]
	}
	t := template.Must(template.New(name).Parse(tmpl))

	return t.Execute(
		out,
		map[string]interface{}{
			"appname":   opts.appname,
			"bin":       opts.dirs.bin,
			"builders":  defaultBuilders(),
			"platforms": defaultPlatforms(),
		},
	)
}

func usagelogo(out *os.File) { //nolint:forbidigo
	_, _ = color.New(color.FgCyan).Fprint(colorable.NewColorable(out), logo)
}

const (
	logo = ` _    __     
| |__/ /__ __
| / / _ \ \ /
|_\_\___/_\_\
`

	otherUsage = `
Launcher Commands:
  deps    Print k6 and extension dependencies
  build   Build custom k6 binary with extensions
  service Start k6x builder service
  preload Preload (go) build cache

Launcher Flags:
  --bin-dir path     cache folder for k6 binary (default: {{.bin}})
  --cache-dir path   set cache base directory
  --with dependency  additional dependency and version constraints
  --filter expr      jmespath syntax extension registry filter (default: [*])
  --builder list     comma separated list of builders (default: {{.builders}})
  --clean            remove cached k6 binary
  --dry              do not run k6 command
`
)
