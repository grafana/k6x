// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

//nolint:revive
package builder

import (
	"context"
	"fmt"
	"io"

	"github.com/sirupsen/logrus"
	"github.com/szkiba/k6x/internal/dependency"
)

func Preload(
	ctx context.Context,
	builder Builder,
	mods dependency.Modules,
	platforms []*Platform,
) error {
	if builder.Engine() == Service {
		return fmt.Errorf("%w: builder service preload not supported", errNoBuilder)
	}

	for _, platform := range platforms {
		logrus.Infof("preloading for %s", platform.String())

		err := builder.Build(ctx, platform, mods, io.Discard)
		if err != nil {
			return err
		}
	}

	return nil
}
