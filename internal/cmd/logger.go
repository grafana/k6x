// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

package cmd

import (
	"io"

	"github.com/briandowns/spinner"
	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
	"github.com/sirupsen/logrus"
)

func initLogger(opts *options) {
	level := logrus.InfoLevel

	if opts.verbose {
		level = logrus.DebugLevel
	}

	if opts.quiet {
		level = logrus.WarnLevel
	}

	logrus.SetLevel(level)

	out := opts.spinner.WriterFile

	if !isatty.IsTerminal(out.Fd()) {
		logrus.SetFormatter(&logrus.JSONFormatter{DisableHTMLEscape: true})
		logrus.SetOutput(out)

		return
	}

	log := newLogFormatter(opts.spinner)
	logrus.SetFormatter(log)
	logrus.AddHook(log)
	logrus.SetOutput(log.out)
}

type logFormatter struct {
	impl    logrus.Formatter
	spinner *spinner.Spinner
	out     io.Writer
}

func newLogFormatter(spinner *spinner.Spinner) *logFormatter {
	f := new(logFormatter)

	f.impl = &logrus.TextFormatter{ForceColors: true}
	f.spinner = spinner
	f.out = colorable.NewColorable(spinner.WriterFile)

	return f
}

func (f *logFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	if entry.Level == logrus.InfoLevel && f.spinner.Enabled() {
		return nil, nil
	}

	return f.impl.Format(entry)
}

func (f *logFormatter) Levels() []logrus.Level {
	return []logrus.Level{logrus.InfoLevel}
}

func (f *logFormatter) Fire(entry *logrus.Entry) error {
	if !f.spinner.Enabled() {
		return nil
	}

	if f.spinner.Active() {
		f.spinner.Lock()
	}

	f.spinner.Suffix = " " + entry.Message

	if f.spinner.Active() {
		f.spinner.Unlock()
	} else {
		f.spinner.Start()
	}

	return nil
}
