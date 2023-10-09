// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

package cmd

import (
	"context"
	"os"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/szkiba/k6x/internal/builder"
	"github.com/szkiba/k6x/internal/dependency"
	"github.com/szkiba/k6x/internal/resolver"
)

func prepare(
	ctx context.Context,
	cmd string,
	deps dependency.Dependencies,
	res resolver.Resolver,
	opts *options,
) error {
	if opts.clean && exists(cmd, opts.dirs.fs) {
		if err := opts.dirs.fs.Remove(cmd); err != nil {
			return err
		}
	}

	var fulfill bool

	if exists(cmd, opts.dirs.fs) {
		cmdres := resolver.FromCommand(cmd, "version")
		bings, err := cmdres.Resolve(ctx, deps)
		if err != nil {
			return err
		}

		fulfill = bings.Resolves(deps)
	}

	if fulfill {
		return nil
	}

	if exists(cmd, opts.dirs.fs) {
		cmdeps, err := resolver.CommandDependencies(ctx, cmd, "version")
		if err != nil {
			return err
		}

		addOptional(ctx, res, deps, cmdeps)
	}

	return build(ctx, deps, res, opts)
}

func addOptional(ctx context.Context, res resolver.Resolver, deps, opt dependency.Dependencies) {
	if len(opt) == 0 {
		return
	}

	ings, _ := res.Resolve(ctx, opt)

	for name, dep := range opt {
		_, has := deps[name]
		_, resolvable := ings[name]

		if !has && resolvable {
			deps[name] = dep
		}
	}
}

func ensureK6(deps dependency.Dependencies) {
	if _, has := deps.K6(); !has {
		deps["k6"] = &dependency.Dependency{Name: "k6"}
	}
}

func addDeps(
	ctx context.Context,
	res resolver.Resolver,
	deps, req dependency.Dependencies,
) error {
	if len(req) == 0 {
		return nil
	}

	_, err := res.Resolve(ctx, req)
	if err != nil {
		return err
	}

	for name, dep := range req {
		deps[name] = dep
	}

	return nil
}

func collectDependencies(
	ctx context.Context,
	res resolver.Resolver,
	opts *options,
) (dependency.Dependencies, error) {
	script := opts.script()
	if script == "-" {
		return nil, errStdinNotSupported
	}

	logrus.Info("search for dependencies")

	deps := make(dependency.Dependencies)

	if len(script) > 0 {
		sdeps, err := dependency.FromScript(script, opts.dirs.fs, deps)
		if err != nil {
			return nil, err
		}

		deps = sdeps
	}

	addOptional(ctx, res, deps, opts.dependencies())

	if err := addDeps(ctx, res, deps, opts.with); err != nil {
		return nil, err
	}

	return deps, nil
}

func build(
	ctx context.Context,
	deps dependency.Dependencies,
	res resolver.Resolver,
	opts *options,
) error {
	ensureK6(deps)

	logrus.Info("resolving dependencies")

	mods, err := res.Resolve(ctx, deps)
	if err != nil {
		return err
	}

	b, err := builder.New(ctx, opts.engines...)
	if err != nil {
		return err
	}

	logrus.Infof("installing k6 (builder: %s, target: %s)", b.Engine().String(), opts.dirs.bin)

	afs := opts.dirs.fs

	if err = afs.MkdirAll(opts.dirs.bin, 0o750); err != nil {
		return err
	}

	fname := filepath.Join(opts.dirs.bin, "k6")
	if runtime.GOOS == "windows" {
		fname += ".exe"
	}

	var file afero.File

	file, err = afs.OpenFile(fname, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755) //nolint:forbidigo
	if err != nil {
		return err
	}

	defer deferredClose(file, &err)

	err = b.Build(ctx, nil, mods, file)
	if err != nil {
		file.Close()      //nolint:gosec,errcheck
		afs.Remove(fname) //nolint:gosec,errcheck

		return err
	}

	return file.Close()
}

func exists(file string, afs afero.Fs) bool {
	_, err := afs.Stat(file)

	return err == nil
}
