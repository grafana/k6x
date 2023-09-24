// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

// Package resolver contains dependency resolving types and interfaces.
//
//nolint:revive
package resolver

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/szkiba/k6x/internal/dependency"
)

const (
	latestTag        = "latest"
	versionTagPrefix = "v"
)

var (
	ErrResolver = errors.New("resolver error")

	k6DefaultIngredient = &Ingredient{ //nolint:gochecknoglobals
		Name:   k6,
		Module: "github.com/grafana/k6",
	}
)

type Ingredient struct {
	Name    string          `json:"name,omitempty"`
	Version *semver.Version `json:"version,omitempty"`
	Module  string          `json:"module,omitempty"`
}

func NewIngredient(name, version, module string) (*Ingredient, error) {
	var err error
	ing := new(Ingredient)

	ing.Name = name
	ing.Module = module

	if len(version) != 0 {
		ing.Version, err = semver.NewVersion(version)
		if err != nil {
			return nil, err
		}
	}

	return ing, nil
}

func (ing *Ingredient) Tag() string {
	if ing.Version == nil {
		return latestTag
	}

	return versionTagPrefix + ing.Version.String()
}

func (ing *Ingredient) String() string {
	var buff strings.Builder

	buff.WriteString(ing.Name)

	if len(ing.Module) != 0 {
		buff.WriteRune(' ')
		buff.WriteString(ing.Module)
		buff.WriteRune('@')

		if ing.Version != nil {
			buff.WriteRune('v')
			buff.WriteString(ing.Version.String())
		} else {
			buff.WriteString("latest")
		}
	}

	return buff.String()
}

type Ingredients map[string]*Ingredient

func (ings Ingredients) Resolves(deps dependency.Dependencies) bool {
	for _, dep := range deps {
		ing, ok := ings[dep.Name]
		if !ok {
			return false
		}

		if !dep.Check(ing.Version) {
			return false
		}
	}

	return true
}

func (ings Ingredients) K6() (*Ingredient, bool) {
	ing, ok := ings[k6]

	return ing, ok
}

func (ings Ingredients) Extensions() []*Ingredient {
	var exts []*Ingredient

	for name, dep := range ings {
		if name != k6 {
			exts = append(exts, dep)
		}
	}

	sort.Slice(exts, func(i, j int) bool {
		return exts[i].Name < exts[j].Name
	})

	return exts
}

func (ings Ingredients) String() string {
	var buff strings.Builder

	if ing, ok := ings.K6(); ok {
		buff.WriteString(ing.String())
		buff.WriteRune('\n')
	}

	for _, ing := range ings.Extensions() {
		buff.WriteString(ing.String())
		buff.WriteRune('\n')
	}

	return buff.String()
}

func (ings Ingredients) filter(deps dependency.Dependencies) Ingredients {
	found := make(Ingredients)

	for _, dep := range deps {
		if ing, ok := ings[dep.Name]; ok && dep.Check(ing.Version) {
			found[dep.Name] = ing
		}
	}

	return found
}

func (ings Ingredients) MarshalJSON() ([]byte, error) {
	dict := make(map[string]string, len(ings))

	for _, ing := range ings {
		var buff strings.Builder

		if len(ing.Module) != 0 {
			buff.WriteString(ing.Module)
		}

		buff.WriteRune('@')

		if ing.Version != nil {
			buff.WriteRune('v')
			buff.WriteString(ing.Version.String())
		} else {
			buff.WriteString("latest")
		}

		dict[ing.Name] = buff.String()
	}

	return json.Marshal(dict)
}

type Resolver interface {
	Resolve(ctx context.Context, dep dependency.Dependencies) (Ingredients, error)
}
