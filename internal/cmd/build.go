// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

package cmd

import (
	"context"

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

func collectDependencies(
	ctx context.Context,
	res resolver.Resolver,
	opts *options,
) (dependency.Dependencies, error) {
	script := opts.script()
	if script == "-" {
		return nil, errStdinNotSupported
	}

	if sp := opts.spinner; sp.Enabled() {
		sp.Stop()
		sp.Suffix = " checking dependencies of " + script
		sp.Start()

		defer sp.Stop()
	}

	deps := make(dependency.Dependencies)
	/*
		if exists(cmd, opts.dirs.fs) {
			cmdeps, err := resolver.CommandDependencies(ctx, cmd, "version")
			if err != nil {
				return nil, err
			}

			addOptional(ctx, res, deps, cmdeps)
		}
	*/
	if len(script) > 0 {
		sdeps, err := dependency.FromScript(script, opts.dirs.fs, deps)
		if err != nil {
			return nil, err
		}

		deps = sdeps
	}

	addOptional(ctx, res, deps, opts.dependencies())

	return deps, nil
}

func build(
	ctx context.Context,
	deps dependency.Dependencies,
	res resolver.Resolver,
	opts *options,
) error {
	if sp := opts.spinner; sp.Enabled() {
		sp.Stop()
		sp.Suffix = " building k6 to " + opts.dirs.bin
		sp.Start()

		defer sp.Stop()
	}

	ings, err := res.Resolve(ctx, deps)
	if err != nil {
		return err
	}

	err = builder.Build(ctx, ings, opts.dirs.bin, opts.dirs.fs)
	if err != nil {
		return err
	}

	return nil
}

func exists(file string, afs afero.Fs) bool {
	_, err := afs.Stat(file)

	return err == nil
}
