// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

// Package dependency contains k6 dependency related types.
//
//nolint:revive
package dependency

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
)

var ErrInvalidConstraints = errors.New("invalid constraints")

type Dependency struct {
	Name        string              `json:"name,omitempty"`
	Constraints *semver.Constraints `json:"constraints,omitempty"`
}

func New(name, constraints string) (*Dependency, error) {
	var err error

	dep := new(Dependency)

	dep.Name = name

	if len(constraints) != 0 {
		if dep.Constraints, err = semver.NewConstraint(constraints); err != nil {
			return nil, fmt.Errorf("%w: %s", ErrInvalidConstraints, err.Error())
		}
	}

	return dep, nil
}

func (dep *Dependency) Check(version *semver.Version) bool {
	return version != nil && (dep.Constraints == nil || dep.Constraints.Check(version))
}

func (dep *Dependency) String() string {
	var buff strings.Builder

	buff.WriteString(dep.Name)
	buff.WriteRune(' ')

	if dep.Constraints != nil {
		buff.WriteString(dep.Constraints.String())
	} else {
		buff.WriteRune('*')
	}

	return buff.String()
}

func (dep *Dependency) update(from *Dependency) error {
	if from.Constraints != nil {
		if dep.Constraints == nil {
			dep.Constraints = from.Constraints
		} else if from.Constraints.String() != dep.Constraints.String() {
			return fmt.Errorf("%w: constraints conflict: %s <-> %s", ErrScript, dep, from)
		}
	}

	return nil
}

type Dependencies map[string]*Dependency

func (deps Dependencies) update(from *Dependency) error {
	dep, found := deps[from.Name]
	if !found {
		deps[from.Name] = from

		return nil
	}

	return dep.update(from)
}

func (deps Dependencies) ensure(from Dependencies) {
	for name := range from {
		if _, found := deps[name]; !found {
			deps[name] = &Dependency{Name: name}
		}
	}
}

func (deps Dependencies) K6() (*Dependency, bool) {
	dep, ok := deps[k6]

	return dep, ok
}

func (deps Dependencies) Extensions() []*Dependency {
	var exts []*Dependency

	for name, dep := range deps {
		if name != k6 {
			exts = append(exts, dep)
		}
	}

	sort.Slice(exts, func(i, j int) bool {
		return exts[i].Name < exts[j].Name
	})

	return exts
}

func (deps Dependencies) String() string {
	var buff strings.Builder

	if dep, ok := deps.K6(); ok {
		buff.WriteString(dep.String())
		buff.WriteRune('\n')
	}

	for _, dep := range deps.Extensions() {
		buff.WriteString(dep.String())
		buff.WriteRune('\n')
	}

	return buff.String()
}

func (deps Dependencies) MarshalJSON() ([]byte, error) {
	dict := make(map[string]string, len(deps))

	for _, dep := range deps {
		constraints := "*"

		if dep.Constraints != nil {
			constraints = dep.Constraints.String()
		}

		dict[dep.Name] = constraints
	}

	var buff bytes.Buffer

	encoder := json.NewEncoder(&buff)

	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(dict); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

const k6 = "k6"
