// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

//nolint:revive
package resolver

import (
	"context"
	"errors"

	"github.com/szkiba/k6x/internal/dependency"
)

var ErrResolver = errors.New("resolver error")

type Resolver interface {
	Resolve(ctx context.Context, dep dependency.Dependencies) (dependency.Modules, error)
	Starred(ctx context.Context, stars int) (dependency.Modules, error)
}
