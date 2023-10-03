// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

// Package builder contains k6 builder logic.
//
//nolint:revive
package builder

import (
	"context"
	"errors"
	"fmt"
	"io"
	"runtime"

	"github.com/szkiba/k6x/internal/dependency"
)

var errNoEngine = errors.New("no suitable builder")

type GOEnv struct {
	GOOS   string
	GOARCH string
	CGO    bool
}

func (g *GOEnv) FromRuntime() *GOEnv {
	g.GOOS = runtime.GOOS
	g.GOARCH = runtime.GOARCH
	g.CGO = false

	return g
}

type Builder interface {
	Build(ctx context.Context, goenv *GOEnv, mods dependency.Modules, out io.Writer) error
}

func New(engines ...Engine) (Builder, error) {
	engs := engines
	if len(engs) == 0 {
		engs = append(engs, Native, Docker)
	}

	for _, eng := range engines {
		if eng == Docker {
			return newDockerBuilder()
		}

		if eng == Native {
			if _, hasGo := goVersion(); hasGo && hasGit() {
				return newNativeBuilder()
			}
		}
	}

	return nil, fmt.Errorf("%w in: %v", errNoEngine, engs)
}

func NewWithType(engine Engine) (Builder, error) {
	switch engine {
	case Docker:
		return newDockerBuilder()
	case Native:
		return newNativeBuilder()
	default:
		panic("unknown builder engine")
	}
}
