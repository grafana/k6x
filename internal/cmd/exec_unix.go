// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

//go:build !windows

package cmd

import (
	"os"

	"golang.org/x/sys/unix"
)

const k6Binary = "k6"

//nolint:forbidigo
func exec(cmd string, args []string, _, _, _ *os.File) (int, error) {
	if err := unix.Exec(cmd, args, os.Environ()); err != nil {
		return exitErr, err
	}

	return 0, nil
}
