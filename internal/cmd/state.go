package cmd

import (
	"context"
	"log/slog"
	"net/url"
	"os"
	"os/exec"

	"github.com/grafana/k6deps"
	"github.com/grafana/k6exec"
	"github.com/spf13/cobra"
)

type state struct {
	k6exec.Options
	buildServiceURL     string
	extensionCatalogURL string
	verbose             bool
	quiet               bool
	nocolor             bool
	usage               bool
	levelVar            *slog.LevelVar

	cmd *exec.Cmd
}

//nolint:forbidigo
func newState(levelVar *slog.LevelVar) *state {
	s := new(state)

	s.levelVar = levelVar

	if value, found := os.LookupEnv("K6_BUILD_SERVICE_URL"); found {
		s.buildServiceURL = value
	}

	if value, found := os.LookupEnv("K6_EXTENSION_CATALOG_URL"); found {
		s.extensionCatalogURL = value
	}

	return s
}

func (s *state) persistentPreRunE(_ *cobra.Command, _ []string) error {
	if len(s.buildServiceURL) > 0 {
		val, err := url.Parse(s.buildServiceURL)
		if err != nil {
			return err
		}

		s.Options.BuildServiceURL = val
	}

	if len(s.extensionCatalogURL) > 0 {
		val, err := url.Parse(s.extensionCatalogURL)
		if err != nil {
			return err
		}

		s.Options.ExtensionCatalogURL = val
	}

	if s.verbose && s.levelVar != nil {
		s.levelVar.Set(slog.LevelDebug)
	}

	return nil
}

func (s *state) preRunE(sub *cobra.Command, args []string) error {
	var (
		deps  k6deps.Dependencies
		err   error
		dopts k6deps.Options
	)

	if scriptname, hasScript := scriptArg(sub, args); hasScript {
		dopts.Script.Name = scriptname
	}

	deps, err = k6deps.Analyze(&dopts)
	if err != nil {
		return err
	}

	cmdargs := make([]string, 0, len(args))

	if sub.Name() != s.Options.AppName {
		cmdargs = append(cmdargs, sub.Name())
	}

	if s.verbose {
		cmdargs = append(cmdargs, "-v")
	}

	if s.quiet {
		cmdargs = append(cmdargs, "-q")
	}

	if s.nocolor {
		cmdargs = append(cmdargs, "--no-color")
	}

	cmdargs = append(cmdargs, args...)

	cmd, err := k6exec.Command(context.Background(), cmdargs, deps, &s.Options)
	if err != nil {
		return err
	}

	cmd.Stderr = os.Stderr //nolint:forbidigo
	cmd.Stdout = os.Stdout //nolint:forbidigo
	cmd.Stdin = os.Stdin   //nolint:forbidigo

	s.cmd = cmd

	return nil
}

func (s *state) runE(_ *cobra.Command, _ []string) error {
	defer k6exec.CleanupState(&s.Options) //nolint:errcheck

	return s.cmd.Run()
}

func (s *state) helpFunc(cmd *cobra.Command, args []string) {
	err := s.preRunE(cmd, append(args, "-h"))
	if err != nil {
		cmd.PrintErr(err)
	}

	err = s.runE(cmd, args)
	if err != nil {
		cmd.PrintErr(err)
	}
}
