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

	"github.com/spf13/afero"
	"github.com/szkiba/k6x/internal/resolver"
)

var errNoEngine = errors.New("no suitable builder")

type Builder interface {
	Build(ctx context.Context, ings resolver.Ingredients, dir string, afs afero.Fs) error
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
