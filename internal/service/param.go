// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

//nolint:revive
package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/szkiba/k6x/internal/builder"
	"github.com/szkiba/k6x/internal/dependency"
	"github.com/szkiba/k6x/internal/resolver"
)

var errInvalidParameters = errors.New("invalid parameters")

type Params struct {
	dependency.Artifacts
	*builder.Platform
}

func platformFromPath(str string) (*builder.Platform, string, error) {
	parts := strings.SplitN(str, "/", 4)
	if len(parts) != 4 {
		return nil, "", errInvalidParameters
	}

	platform := builder.NewPlatform(parts[1], parts[2])

	if !platform.Supported() {
		return nil, "", errUnsupportedPlatform
	}

	return platform, parts[3], nil
}

func parseParams(str string) (*Params, error) {
	platform, deplist, err := platformFromPath(str)
	if err != nil {
		return nil, err
	}

	arts, err := dependency.ParseArtifacts(deplist)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errInvalidParameters, err.Error())
	}

	if _, has := arts.K6(); !has {
		return nil, fmt.Errorf("%w: missing k6 parameter", errInvalidParameters)
	}

	return &Params{Artifacts: arts, Platform: platform}, nil
}

func looseParseParams(ctx context.Context, str string, res resolver.Resolver) (*Params, error) {
	platform, deplist, err := platformFromPath(str)
	if err != nil {
		return nil, err
	}

	deps, err := dependency.ParseLooseArtifacts(deplist)
	if err != nil {
		return nil, err
	}

	if _, has := deps.K6(); !has {
		deps["k6"] = &dependency.Dependency{Name: "k6"}
	}

	mods, err := res.Resolve(ctx, deps)
	if err != nil {
		return nil, errInvalidParameters
	}

	return &Params{Artifacts: mods.ToArtifacts(), Platform: platform}, nil
}

func (pars *Params) String() string {
	var buff strings.Builder

	buff.WriteRune('/')
	buff.WriteString(pars.Platform.String())
	buff.WriteRune('/')
	buff.WriteString(pars.Artifacts.String())

	return buff.String()
}

func (pars *Params) ETag() string {
	sum := sha256.Sum256([]byte(pars.String()))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

var errUnsupportedPlatform = errors.New("unsupported platform")
