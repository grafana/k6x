// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

//go:build windows

package cmd

import "os"

const k6Binary = "k6.exe"

//nolint:forbidigo
func exec(cmd string, args []string, stdin, stdout, stderr *os.File) (int, error) {
	proc, err := os.StartProcess(cmd, args, &os.ProcAttr{
		Files: []*os.File{stdin, stdout, stderr},
	})
	if err != nil {
		return exitErr, err
	}

	st, err := proc.Wait()
	if err != nil {
		return exitErr, err
	}

	return st.ExitCode(), nil
}
