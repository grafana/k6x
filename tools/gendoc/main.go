// Package main contains CLI documentation generator tool.
package main

import (
	_ "embed"
	"strings"

	"github.com/grafana/clireadme"
	"github.com/grafana/k6x/internal/cmd"
)

func main() {
	root := cmd.New(nil)
	root.Use = strings.ReplaceAll(root.Use, "exec", "k6x")
	clireadme.Main(root, 1)
}
