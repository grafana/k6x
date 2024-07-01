// Package cmd contains run cobra command factory function.
package cmd

import (
	_ "embed"
	"log/slog"

	"github.com/spf13/cobra"
)

//go:embed help.md
var help string

// New creates new cobra command for exec command.
func New(levelVar *slog.LevelVar) *cobra.Command {
	state := newState(levelVar)

	root := &cobra.Command{
		Use:               "exec [flags] [command]",
		Short:             "Lanch k6 with extensions",
		Long:              help,
		SilenceUsage:      true,
		SilenceErrors:     true,
		DisableAutoGenTag: true,
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if state.usage {
				return nil
			}

			state.AppName = cmd.Name()

			return state.preRunE(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.usage {
				return cmd.Help()
			}

			return state.runE(cmd, args)
		},
		PersistentPreRunE: state.persistentPreRunE,
	}

	root.SetVersionTemplate(`{{with .Name}}{{printf "%s " .}}{{end}}{{printf "%s\n" .Version}}`)

	for _, name := range commands {
		root.AddCommand(newSubcommand(name, state))
	}

	flags := root.PersistentFlags()

	flags.StringVar(
		&state.extensionCatalogURL,
		"extension-catalog-url",
		state.extensionCatalogURL,
		"URL of the k6 extension catalog to be used",
	)
	flags.StringVar(
		&state.buildServiceURL,
		"build-service-url",
		state.buildServiceURL,
		"URL of the k6 build service to be used",
	)
	flags.BoolVarP(&state.verbose, "verbose", "v", false, "enable verbose logging")
	flags.BoolVarP(&state.quiet, "quiet", "q", false, "disable progress updates")
	flags.BoolVar(&state.quiet, "no-color", false, "disable colored output")
	flags.BoolVar(&state.usage, "usage", false, "print launcher usage")

	root.InitDefaultHelpFlag()
	root.Flags().Lookup("help").Usage = "help for k6"

	root.MarkFlagsMutuallyExclusive("extension-catalog-url", "build-service-url")

	return root
}

func scriptArg(cmd *cobra.Command, args []string) (string, bool) {
	if len(cmd.Annotations) == 0 {
		return "", false
	}

	if _, use := cmd.Annotations[useExtensions]; !use {
		return "", false
	}

	if len(args) == 0 {
		return "", false
	}

	last := args[len(args)-1]
	if len(last) == 0 || last[0] == '-' {
		return "", false
	}

	return last, true
}

func newSubcommand(name string, state *state) *cobra.Command {
	cmd := &cobra.Command{
		Use:                name,
		PreRunE:            state.preRunE,
		RunE:               state.runE,
		SilenceErrors:      true,
		SilenceUsage:       true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		Hidden:             true,
	}
	cmd.SetHelpFunc(state.helpFunc)

	if name == "run" || name == "archive" {
		cmd.Annotations = map[string]string{useExtensions: "true"}
	}

	return cmd
}

const useExtensions = "useExtensions"

var commands = []string{ //nolint:gochecknoglobals
	"help",
	"resume",
	"scale",
	"cloud",
	"completion",
	"inspect",
	"pause",
	"status",
	"login",
	"stats",
	"version",
	"new",
	"run",
	"archive",
}
