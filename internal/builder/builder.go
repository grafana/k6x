// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

// Package builder contains k6 builder logic.
//
//nolint:revive
package builder

import (
	"context"
	"fmt"
	"io"

	"github.com/szkiba/k6x/internal/dependency"
)

type Builder interface {
	Build(ctx context.Context, platform *Platform, mods dependency.Modules, out io.Writer) error
	Engine() Engine
}

func New(ctx context.Context, engines ...Engine) (Builder, error) {
	engs := engines
	if len(engs) == 0 {
		engs = append(engs, DefaultEngines()...)
	}

	for _, eng := range engines {
		impl, found, err := eng.NewBuilder(ctx)
		if err != nil {
			return nil, err
		}

		if found {
			return impl, nil
		}
	}

	return nil, fmt.Errorf("%w in: %v", errNoBuilder, engs)
}
