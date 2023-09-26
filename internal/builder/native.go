// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

//nolint:revive,forbidigo
package builder

import (
	"bytes"
	"context"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/szkiba/k6x/internal/resolver"
	"go.k6.io/xk6"
)

type nativeBuilder struct {
	stderr    *os.File
	logWriter *io.PipeWriter
	logFlags  int
	logOutput io.Writer
}

func goVersion() (*semver.Version, bool) {
	cmd, err := exec.LookPath("go")
	if err != nil {
		return nil, false
	}

	out, err := exec.Command(cmd, "version").Output() //nolint:gosec
	if err != nil {
		return nil, false
	}

	pre := []byte("go")

	fields := bytes.SplitN(out, []byte{' '}, 4)
	if len(fields) < 4 || !bytes.Equal(fields[0], pre) || !bytes.HasPrefix(fields[2], pre) {
		return nil, false
	}

	ver, err := semver.NewVersion(string(bytes.TrimPrefix(fields[2], pre)))
	if err != nil {
		return nil, false
	}

	return ver, true
}

func hasGit() bool {
	cmd, err := exec.LookPath("git")
	if err != nil {
		return false
	}

	_, err = exec.Command(cmd, "version").Output() //nolint:gosec

	return err == nil
}

func newNativeBuilder() (*nativeBuilder, error) {
	return new(nativeBuilder), nil
}

func (b *nativeBuilder) Build(
	ctx context.Context,
	ings resolver.Ingredients,
	dir string,
	afs afero.Fs,
) error {
	b.logFlags = log.Flags()
	b.logOutput = log.Writer()
	b.logWriter = logrus.StandardLogger().WriterLevel(logrus.DebugLevel)
	b.stderr = os.Stderr

	log.SetOutput(b.logWriter)
	log.SetFlags(0)

	if null, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0); err == nil {
		os.Stderr = null
	}

	defer b.close()

	return b.build(ctx, ings, dir, afs)
}

func (b *nativeBuilder) close() {
	_ = b.logWriter.Close()

	log.SetFlags(b.logFlags)
	log.SetOutput(b.logOutput)

	os.Stderr = b.stderr
}

func (b *nativeBuilder) build(
	ctx context.Context,
	ings resolver.Ingredients,
	dir string,
	afs afero.Fs,
) error {
	logrus.Debug("Building new k6 binary (native)")

	if err := afs.MkdirAll(dir, 0o750); err != nil {
		return err
	}

	builder := new(xk6.Builder)

	if k6, has := ings.K6(); has {
		builder.K6Version = k6.Tag()
	}

	for _, ing := range ings.Extensions() {
		builder.Extensions = append(builder.Extensions,
			xk6.Dependency{
				PackagePath: ing.Module,
				Version:     ing.Tag(),
			},
		)
	}

	return builder.Build(ctx, filepath.Join(dir, "k6"))
}
