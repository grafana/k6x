// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

package cmd

import (
	"context"
	"os"

	"github.com/szkiba/k6x/internal/resolver"
)

func otherCommand(
	ctx context.Context,
	cmd string,
	res resolver.Resolver,
	opts *options,
	stdin, stdout, stderr *os.File, //nolint:forbidigo
) (int, error) {
	if err := prepare(ctx, cmd, nil, res, opts); err != nil {
		return exitErr, err
	}

	if opts.dry {
		return 0, nil
	}

	opts.argv[0] = cmd

	return exec(cmd, opts.argv, stdin, stdout, stderr)
}
