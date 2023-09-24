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

	initLogger(opts)

	res := resolver.NewWithCacheDir(opts.dirs.http)
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

	if opts.help || len(opts.argv) == 1 {
		usagelogo(stdout)

		if err := usage(stdout, otherUsage, opts); err != nil {
			return exitErr, err
		}
	}

	if opts.run() {
		return runCommand(ctx, cmd, res, opts, stdin, stdout, stderr)
	}

	if opts.version() {
		return versionCommand(ctx, cmd, res, opts, stdin, stdout, stderr)
	}

	return otherCommand(ctx, cmd, res, opts, stdin, stdout, stderr)
}

func initLogger(opts *options) {
	level := logrus.InfoLevel

	if opts.verbose {
		level = logrus.DebugLevel
	}

	if opts.quiet {
		level = logrus.WarnLevel
	}

	logrus.SetLevel(level)

	logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})
	logrus.SetOutput(colorable.NewColorableStdout())
}

func usage(out io.Writer, tmpl string, opts *options) error {
	t := template.Must(template.New("usage").Parse(tmpl))

	return t.Execute(out, map[string]interface{}{"appname": opts.appname, "bin": opts.dirs.bin})
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
  deps   Print k6 and extension dependencies
  build  Build custom k6 binary with extensions

Launcher Flags:
  --bin-dir path  cache folder for k6 binary (default: {{.bin}})
  --clean         remove cached k6 binary
  --dry           do not run k6 command
`
)
