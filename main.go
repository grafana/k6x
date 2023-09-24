// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

//nolint:forbidigo
package main

import (
	"context"
	"os"

	"github.com/spf13/afero"
	"github.com/szkiba/k6x/internal/cmd"
)

func main() {
	os.Exit(cmd.Main(context.TODO(), os.Args, os.Stdin, os.Stdout, os.Stderr, afero.NewOsFs()))
}
