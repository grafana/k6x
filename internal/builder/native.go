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

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
	"github.com/szkiba/k6x/internal/dependency"
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

func newNativeBuilder(_ context.Context) (Builder, bool, error) {
	if _, hasGo := goVersion(); !hasGo || !hasGit() {
		return nil, false, nil
	}

	return new(nativeBuilder), true, nil
}

func (b *nativeBuilder) Engine() Engine {
	return Native
}

func (b *nativeBuilder) Build(
	ctx context.Context,
	platform *Platform,
	mods dependency.Modules,
	out io.Writer,
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

	if platform == nil {
		platform = RuntimePlatform()
	}

	return b.build(ctx, platform, mods, out)
}

func (b *nativeBuilder) close() {
	_ = b.logWriter.Close()

	log.SetFlags(b.logFlags)
	log.SetOutput(b.logOutput)

	os.Stderr = b.stderr
}

func (b *nativeBuilder) build(
	ctx context.Context,
	platform *Platform,
	mods dependency.Modules,
	out io.Writer,
) error {
	logrus.Debug("Building new k6 binary (native)")

	builder := new(xk6.Builder)

	builder.Cgo = false
	builder.OS = platform.OS
	builder.Arch = platform.Arch
	builder.Replacements = newReplacements(mods, replacementsFromContext(ctx))

	if k6, has := mods.K6(); has {
		builder.K6Version = k6.Tag()
	}

	for _, ing := range mods.Extensions() {
		builder.Extensions = append(builder.Extensions,
			xk6.Dependency{
				PackagePath: ing.Path,
				Version:     ing.Tag(),
			},
		)
	}

	tmp, err := os.CreateTemp("", "k6")
	if err != nil {
		return err
	}

	if err = tmp.Close(); err != nil {
		return err
	}

	if err = builder.Build(ctx, tmp.Name()); err != nil {
		return err
	}

	tmp, err = os.Open(tmp.Name())
	if err != nil {
		return err
	}

	_, err = io.Copy(out, tmp)

	tmp.Close()           //nolint:errcheck,gosec
	os.Remove(tmp.Name()) //nolint:errcheck,gosec

	return err
}

func newReplacements(mods dependency.Modules, reps Replacements) []xk6.Replace {
	replacements := make([]xk6.Replace, 0, len(reps))

	for name, mod := range mods {
		if rep, has := reps[name]; has {
			old := xk6.ReplacementPath(mod.Path)
			to := xk6.ReplacementPath(rep.Path)
			replacements = append(replacements, xk6.Replace{Old: old, New: to})
		}
	}

	return replacements
}
