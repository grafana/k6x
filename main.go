// Package main contains the main function for k6x.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/grafana/k6x/internal/cmd"
	sloglogrus "github.com/samber/slog-logrus/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

//nolint:gochecknoglobals
var (
	appname = "k6x"
	version = "dev"
)

func initLogging(app string) *slog.LevelVar {
	levelVar := new(slog.LevelVar)

	logrus.SetLevel(logrus.DebugLevel)

	logger := slog.New(sloglogrus.Option{Level: levelVar}.NewLogrusHandler())
	logger = logger.With("app", app)

	slog.SetDefault(logger)

	return levelVar
}

func main() {
	runCmd(newCmd(os.Args[1:], initLogging(appname))) //nolint:forbidigo
}

func newCmd(args []string, levelVar *slog.LevelVar) *cobra.Command {
	cmd := cmd.New(levelVar)
	cmd.Version = version

	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
		args[0] = "help"
	}

	sp := addSpinner(cmd)

	if len(args) == 1 && args[0] == "--usage" {
		sp.Disable()
	}

	cmd.SetArgs(args)

	return cmd
}

func runCmd(cmd *cobra.Command) {
	if err := cmd.Execute(); err != nil {
		slog.Error(formatError(err))
		os.Exit(1) //nolint:forbidigo
	}
}

func addSpinner(root *cobra.Command) *spinner.Spinner {
	sp := spinner.New(
		spinner.CharSets[11],
		200*time.Millisecond,
		spinner.WithColor("cyan"),
		spinner.WithWriterFile(os.Stderr), //nolint:forbidigo
	)

	red := color.New(color.FgRed)
	green := color.New(color.FgGreen)

	sp.Prefix = "Preparing k6 "
	sp.FinalMSG = sp.Prefix + green.Sprint("✓") + "\n"

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		sp.FinalMSG = sp.Prefix + red.Sprint("✗") + "\n"

		sp.Stop()
	}()

	if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		sp.Disable()
	}

	prerun := root.PersistentPreRunE
	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if cmd == root && len(args) != 0 {
			return nil
		}

		sp.Start()
		cmdpre := cmd.PreRunE
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
			var err error

			if cmdpre != nil {
				err = cmdpre(cmd, args)
				if err != nil {
					sp.FinalMSG = sp.Prefix + red.Sprint("✗") + "\n"
				}
			}

			sp.Stop()
			return err
		}

		return prerun(cmd, args)
	}

	return sp
}
