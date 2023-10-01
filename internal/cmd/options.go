// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/briandowns/spinner"
	"github.com/spf13/afero"
	"github.com/spf13/pflag"
	"github.com/szkiba/k6x/internal/builder"
	"github.com/szkiba/k6x/internal/dependency"
)

const (
	cmdDeps    = "deps"
	cmdBuild   = "build"
	cmdRun     = "run"
	cmdVersion = "version"
)

type directories struct {
	base string
	bin  string
	http string

	fs afero.Fs
}

type options struct {
	verbose bool
	quiet   bool
	help    bool
	nocolor bool
	resolve bool
	json    bool
	clean   bool
	dry     bool
	engines []builder.Engine
	out     []string
	with    dependency.Dependencies
	args    []string
	argv    []string
	dirs    *directories
	appname string
	spinner *spinner.Spinner
}

func checkargs(args []string, appname string) error {
	if len(args) <= 1 {
		return nil
	}

	cmd := args[1]

	if len(args) > 3 {
		return fmt.Errorf(
			"%s %s: %w: received %d",
			appname,
			cmd,
			errOneArg,
			len(args)-2,
		)
	}

	if len(args) == 3 && args[2] == "-" {
		return fmt.Errorf("%s %s: %w", appname, cmd, errStdinNotSupported)
	}

	return nil
}

func cleanargv(argv []string) []string {
	var clean []string
	for i := 0; i < len(argv); i++ {
		arg := argv[i]
		if arg == "--clean" {
			continue
		}

		if arg == "--bin-dir" || arg == "--builder" || arg == "--with" {
			i++
			continue
		}

		clean = append(clean, arg)
	}

	return clean
}

func newFlagSet(opts *options) *pflag.FlagSet {
	flag := pflag.NewFlagSet("root", pflag.ContinueOnError)

	flag.ParseErrorsWhitelist.UnknownFlags = true

	flag.BoolVarP(&opts.verbose, "verbose", "v", false, "")
	flag.BoolVarP(&opts.quiet, "quiet", "q", false, "")
	flag.BoolVarP(&opts.help, "help", "h", false, "")
	flag.BoolVar(&opts.nocolor, "no-color", false, "")

	flag.StringArrayVarP(&opts.out, "out", "o", []string{}, "")

	// deps command
	flag.BoolVar(&opts.resolve, "resolve", false, "")
	flag.BoolVar(&opts.json, "json", false, "")

	// k6 commands
	flag.BoolVar(&opts.clean, "clean", false, "")
	flag.BoolVar(&opts.dry, "dry", false, "")

	// k6 options without argumens
	var dummy bool

	for _, opt := range k6NoArgOpts {
		flag.BoolVar(&dummy, opt, false, "")
	}

	return flag
}

func parseWith(with []string) (dependency.Dependencies, error) {
	deps := make(dependency.Dependencies)

	for _, with := range with {
		if len(with) == 0 {
			return nil, errInvalidWith
		}

		constraints := ""

		parts := strings.SplitN(with, " ", 2)
		if len(parts) > 1 {
			constraints = parts[1]
		}

		dep, err := dependency.New(parts[0], constraints)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", errInvalidWith, err.Error())
		}

		deps[dep.Name] = dep
	}

	return deps, nil
}

func getopts(argv []string, afs afero.Fs) (*options, error) {
	var err error

	opts := new(options)

	opts.appname = _appname

	opts.dirs, err = getdirs(opts.appname, afs)
	if err != nil {
		return nil, err
	}

	opts.argv = cleanargv(argv)
	opts.args = make([]string, len(argv))
	copy(opts.args, argv)

	flag := newFlagSet(opts)

	engines := os.Getenv(strings.ToUpper(opts.appname) + "_BUILDER") //nolint:forbidigo
	if len(engines) == 0 {
		engines = "native,docker"
	}

	builders := flag.StringSlice("builder", strings.Split(engines, ","), "")
	bin := flag.String("bin-dir", "", "")
	with := flag.StringArray("with", []string{}, "")

	if err = flag.Parse(opts.args); err != nil {
		return nil, err
	}

	opts.args = flag.Args()

	if len(*bin) == 0 {
		if opts.build() {
			opts.dirs.bin = "."
		}
	} else {
		opts.dirs.bin = *bin
	}

	for _, val := range *builders {
		var eng builder.Engine

		eng, err = builder.EngineString(val)
		if err != nil {
			return nil, err
		}

		opts.engines = append(opts.engines, eng)
	}

	if opts.with, err = parseWith(*with); err != nil {
		return nil, err
	}

	opts.spinner = getspinner(opts)

	if opts.help {
		return opts, nil
	}

	if err := checkargs(opts.args, opts.appname); err != nil {
		return nil, err
	}

	for i := range opts.out {
		parts := strings.SplitN(opts.out[i], "=", 2)

		opts.out[i] = parts[0]
	}

	return opts, nil
}

func (opts *options) run() bool {
	return len(opts.args) > 1 && opts.args[1] == cmdRun
}

func (opts *options) deps() bool {
	return len(opts.args) > 1 && opts.args[1] == cmdDeps
}

func (opts *options) build() bool {
	return len(opts.args) > 1 && opts.args[1] == cmdBuild
}

func (opts *options) version() bool {
	return len(opts.args) > 1 && opts.args[1] == cmdVersion
}

func (opts *options) script() string {
	if len(opts.args) < 3 {
		return ""
	}

	return opts.args[2]
}

func (opts *options) dependencies() dependency.Dependencies {
	deps := make(dependency.Dependencies)

	for _, output := range opts.out {
		parts := strings.SplitN(output, "=", 2)
		deps[parts[0]] = &dependency.Dependency{Name: parts[0]}
	}

	return deps
}

func bindir(appname string, basedir string, afs afero.Fs) string {
	dir := os.Getenv(strings.ToUpper(appname) + "_BIN_DIR") //nolint:forbidigo
	if len(dir) == 0 && exists("."+appname, afs) {
		dir = "." + appname
	}

	if len(dir) == 0 {
		dir = filepath.Join(basedir, "bin")
	}

	return dir
}

func getdirs(appname string, afs afero.Fs) (*directories, error) {
	var err error
	dirs := new(directories)

	dirs.fs = afs

	dirs.base = filepath.Join(xdg.CacheHome, appname)
	dirs.http = filepath.Join(dirs.base, "http")
	dirs.bin = bindir(appname, dirs.base, afs)

	if err = afs.MkdirAll(dirs.bin, 0o750); err != nil {
		return nil, err
	}

	if err = afs.MkdirAll(dirs.http, 0o750); err != nil {
		return nil, err
	}

	return dirs, nil
}

func getspinner(opts *options) *spinner.Spinner {
	var sopts []spinner.Option

	if !opts.nocolor {
		sopts = append(sopts, spinner.WithColor("magenta"))
	}

	sp := spinner.New(spinner.CharSets[51], 200*time.Millisecond, sopts...)
	sp.Reverse()

	if opts.verbose || opts.quiet || opts.help {
		sp.Disable()
	}

	return sp
}

var (
	errOneArg            = errors.New("accepts at most 1 arg")
	errStdinNotSupported = errors.New("standard input is not supported")
	errInvalidWith       = errors.New("invalid with flag value")

	k6NoArgOpts = []string{ //nolint:gochecknoglobals
		"no-usage-report",
		"no-summary",
		"no-thresholds",
		"discard-response-bodies",
		"throw", "w",
		"no-vu-connection-reuse",
		"no-connection-reuse",
		"no-teardown",
		"no-setup",
		"paused", "p",
	}
)
