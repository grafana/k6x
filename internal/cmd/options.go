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
	cmdService = "service"
	cmdPreload = "preload"
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
	filter  string
	out     []string
	with    dependency.Dependencies
	reps    builder.Replacements
	addr    string
	args    []string
	argv    []string
	dirs    *directories
	appname string
	spinner *spinner.Spinner

	platforms []*builder.Platform
	stars     int
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

		if arg == "--bin-dir" || arg == "--cache-dir" || arg == "--builder" ||
			arg == "--with" || arg == "--replace" ||
			arg == "--filter" {
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

	filter := os.Getenv(strings.ToUpper(opts.appname) + "_FILTER") //nolint:forbidigo
	if len(filter) == 0 {
		filter = "[*]"
	}

	// directories
	flag.StringVar(&opts.dirs.bin, "bin-dir", "", "")
	flag.StringVar(&opts.dirs.base, "cache-dir", "", "")

	flag.StringVar(&opts.filter, "filter", filter, "")

	// deps command
	flag.BoolVar(&opts.resolve, "resolve", false, "")
	flag.BoolVar(&opts.json, "json", false, "")

	// service command
	flag.StringVar(&opts.addr, "addr", "127.0.0.1:8787", "")

	// preload command
	flag.IntVar(&opts.stars, "stars", defaultStars, "")

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

func parseReplace(replace []string) (builder.Replacements, error) {
	reps := make(builder.Replacements)

	for _, rep := range replace {
		if len(rep) == 0 {
			return nil, errInvalidReplace
		}

		parts := strings.SplitN(rep, "=", 2)
		if len(parts) == 1 {
			return nil, errInvalidReplace
		}

		name, path := parts[0], parts[1]

		if strings.HasPrefix(path, ".") {
			var err error

			path, err = filepath.Abs(path)
			if err != nil {
				return nil, err
			}
		}

		reps[name] = builder.NewReplacement(name, path)
	}

	return reps, nil
}

func parseBuilders(builders []string) ([]builder.Engine, error) {
	all := make([]builder.Engine, 0, len(builders))

	for _, val := range builders {
		eng, err := builder.EngineString(val)
		if err != nil {
			return nil, err
		}

		all = append(all, eng)
	}

	return all, nil
}

func parsePlatforms(platforms []string) ([]*builder.Platform, error) {
	all := make([]*builder.Platform, 0, len(platforms))

	for _, val := range platforms {
		platform, err := builder.ParsePlatform(val)
		if err != nil {
			return nil, err
		}

		all = append(all, platform)
	}

	return all, nil
}

func getopts(argv []string, afs afero.Fs) (*options, error) {
	var err error

	opts := new(options)
	opts.appname = _appname

	opts.dirs = new(directories)
	opts.dirs.fs = afs

	opts.argv = cleanargv(argv)
	opts.args = make([]string, len(argv))
	copy(opts.args, argv)

	flag := newFlagSet(opts)

	engines := os.Getenv(strings.ToUpper(opts.appname) + "_BUILDER") //nolint:forbidigo
	if len(engines) == 0 {
		engines = defaultBuilders()
	}

	builders := flag.StringSlice("builder", strings.Split(engines, ","), "")
	platforms := flag.StringSlice("platform", strings.Split(defaultPlatforms(), ","), "")
	with := flag.StringArray("with", []string{}, "")
	replace := flag.StringArray("replace", []string{}, "")

	if err = flag.Parse(opts.args); err != nil {
		return nil, err
	}

	opts.args = flag.Args()

	err = fixdirs(opts)
	if err != nil {
		return nil, err
	}

	if opts.engines, err = parseBuilders(*builders); err != nil {
		return nil, err
	}

	if opts.platforms, err = parsePlatforms(*platforms); err != nil {
		return nil, err
	}

	if opts.with, err = parseWith(*with); err != nil {
		return nil, err
	}

	if opts.reps, err = parseReplace(*replace); err != nil {
		return nil, err
	}

	if len(opts.reps) > 0 {
		opts.engines = []builder.Engine{builder.Native}
		opts.clean = true
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

func (opts *options) service() bool {
	return len(opts.args) > 1 && opts.args[1] == cmdService
}

func (opts *options) preload() bool {
	return len(opts.args) > 1 && opts.args[1] == cmdPreload
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

func (opts *options) exec(
	cmd string,
	args []string,
	stdin, stdout, stderr *os.File, //nolint:forbidigo
) (int, error) {
	if opts.spinner.Enabled() {
		opts.spinner.Stop()
	}

	if opts.dry {
		return 0, nil
	}

	return exec(cmd, args, stdin, stdout, stderr)
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

func fixdirs(opts *options) error {
	if len(opts.dirs.bin) == 0 && opts.build() {
		opts.dirs.bin = "."
	}

	var err error

	dirs := opts.dirs

	if len(dirs.base) == 0 {
		dir := os.Getenv(strings.ToUpper(opts.appname) + "_CACHE_DIR") //nolint:forbidigo
		if len(dir) != 0 {
			dirs.base = dir
		} else {
			dirs.base = filepath.Join(xdg.CacheHome, opts.appname)
		}
	}

	dirs.http = filepath.Join(dirs.base, "http")

	if len(dirs.bin) == 0 {
		dirs.bin = bindir(opts.appname, dirs.base, dirs.fs)
	}

	if err = dirs.fs.MkdirAll(dirs.bin, 0o750); err != nil {
		return err
	}

	return dirs.fs.MkdirAll(dirs.http, 0o750)
}

//nolint:forbidigo
func getspinner(opts *options) *spinner.Spinner {
	var sopts []spinner.Option

	if !opts.nocolor {
		sopts = append(sopts, spinner.WithColor("magenta"))
	}

	sopts = append(sopts, spinner.WithWriterFile(os.Stderr))

	sp := spinner.New(spinner.CharSets[51], 200*time.Millisecond, sopts...)
	sp.Reverse()
	sp.Prefix = opts.appname + " "

	if opts.verbose || opts.quiet {
		sp.Disable()
	}

	return sp
}

var (
	errOneArg            = errors.New("accepts at most 1 arg")
	errStdinNotSupported = errors.New("standard input is not supported")
	errInvalidWith       = errors.New("invalid with flag value")
	errInvalidReplace    = errors.New("invalid replace flag value")

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

func defaultBuilders() string {
	var all strings.Builder

	for _, eng := range builder.DefaultEngines() {
		if all.Len() != 0 {
			all.WriteRune(',')
		}

		all.WriteString(eng.String())
	}

	return all.String()
}

func defaultPlatforms() string {
	var all strings.Builder

	for _, plat := range builder.SupportedPlatforms() {
		if all.Len() != 0 {
			all.WriteRune(',')
		}

		all.WriteString(plat.String())
	}

	return all.String()
}

const defaultStars = 5
