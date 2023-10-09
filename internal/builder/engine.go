// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

//nolint:revive
package builder

import (
	"context"
	"errors"
)

type Engine int

const (
	Native Engine = iota
	Docker
	Service
)

func DefaultEngines() []Engine {
	return []Engine{Service, Native, Docker}
}

var errNoBuilder = errors.New("no suitable builder")

type builderCtor func(ctx context.Context) (Builder, bool, error)

var builders = map[Engine]builderCtor{ //nolint:gochecknoglobals
	Native:  newNativeBuilder,
	Docker:  newDockerBuilder,
	Service: newServiceBuilder,
}

func (e Engine) NewBuilder(ctx context.Context) (Builder, bool, error) {
	ctor, ok := builders[e]
	if !ok {
		return nil, false, errNoBuilder
	}

	return ctor(ctx)
}

//go:generate enumer -transform kebab -json -text -values -type Engine
