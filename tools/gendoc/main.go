// Package main contains CLI documentation generator tool.
package main

import (
	"strings"

	"github.com/grafana/clireadme"
	"github.com/grafana/k6exec/cmd"
)

func main() {
	root := cmd.New(nil)
	root.Use = strings.ReplaceAll(root.Use, "k6exec", "k6x")
	root.Long = strings.ReplaceAll(root.Long, "k6exec", "k6x")
	clireadme.Main(root, 0)
}
