// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

//nolint:revive
package dependency

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
)

var ErrInvalidArtifact = errors.New("invalid artifact")

type Artifact struct {
	Name    string          `json:"name,omitempty"`
	Version *semver.Version `json:"version,omitempty"`
}

func NewArtifact(name, version string) (*Artifact, error) {
	var err error
	art := new(Artifact)

	art.Name = name

	art.Version, err = semver.NewVersion(version)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidArtifact, err.Error())
	}

	return art, nil
}

func (art *Artifact) UnmarshalText(text []byte) error {
	idx := bytes.IndexRune(text, '@')
	if idx <= 0 {
		return fmt.Errorf("%w: missing '@' %s", ErrInvalidArtifact, string(text))
	}

	art.Name = string(text[:idx])

	var err error

	art.Version, err = semver.NewVersion(string(text[idx+1:]))

	return err
}

func (art *Artifact) MarshalText() ([]byte, error) {
	var buff bytes.Buffer

	buff.WriteString(art.Name)
	buff.WriteRune('@')
	buff.WriteRune('v')
	buff.WriteString(art.Version.String())

	return buff.Bytes(), nil
}

func (art *Artifact) ToDependency() *Dependency {
	constraints, _ := semver.NewConstraint(art.Version.String())

	return &Dependency{Name: art.Name, Constraints: constraints}
}

func ParseArtifact(str string) (*Artifact, error) {
	idx := strings.IndexRune(str, '@')
	if idx <= 0 {
		return nil, fmt.Errorf("%w: missing '@' %s", ErrInvalidArtifact, str)
	}

	return NewArtifact(str[:idx], str[idx+1:])
}

func (art *Artifact) String() string {
	var buff strings.Builder

	buff.WriteString(art.Name)
	buff.WriteRune('@')
	buff.WriteRune('v')
	buff.WriteString(art.Version.String())

	return buff.String()
}

func ParseLooseArtifact(str string) (*Dependency, error) {
	name := str
	constraint := ""

	idx := strings.IndexRune(str, '@')
	if idx >= 0 {
		name = str[:idx]
		constraint = str[idx+1:]
	}

	return New(name, constraint)
}

type Artifacts map[string]*Artifact

func ParseArtifacts(str string) (Artifacts, error) {
	arts := make(Artifacts)
	parts := strings.Split(str, ",")

	for _, part := range parts {
		art, err := ParseArtifact(part)
		if err != nil {
			return nil, err
		}

		arts[art.Name] = art
	}

	return arts, nil
}

func (arts Artifacts) UnmarshalText(text []byte) error {
	parts := bytes.Split(text, []byte{','})

	for _, part := range parts {
		art, err := ParseArtifact(string(part))
		if err != nil {
			return err
		}

		arts[art.Name] = art
	}

	return nil
}

func (arts Artifacts) MarshalText() ([]byte, error) {
	var buff bytes.Buffer

	for _, art := range arts.Sorted() {
		if buff.Len() != 0 {
			buff.WriteRune(',')
		}

		b, err := art.MarshalText()
		if err != nil {
			return nil, err
		}

		buff.Write(b)
	}

	return buff.Bytes(), nil
}

func (arts Artifacts) Sorted() []*Artifact {
	all := make([]*Artifact, 0, len(arts))

	for _, art := range arts {
		all = append(all, art)
	}

	sort.Slice(all, func(i, j int) bool {
		if all[i].Name == k6 {
			return true
		}

		if all[j].Name == k6 {
			return false
		}

		return all[i].Name < all[j].Name
	})

	return all
}

func (arts Artifacts) K6() (*Artifact, bool) {
	art, ok := arts[k6]

	return art, ok
}

func (arts Artifacts) String() string {
	var buff strings.Builder

	for _, art := range arts.Sorted() {
		if buff.Len() != 0 {
			buff.WriteRune(',')
		}

		buff.WriteString(art.String())
	}

	return buff.String()
}

func (arts Artifacts) ToDependencies() Dependencies {
	deps := make(Dependencies, len(arts))

	for _, art := range arts {
		dep := art.ToDependency()

		deps[dep.Name] = dep
	}

	return deps
}

func ParseLooseArtifacts(str string) (Dependencies, error) {
	deps := make(Dependencies)
	parts := strings.Split(str, ",")

	for _, part := range parts {
		dep, err := ParseLooseArtifact(part)
		if err != nil {
			return nil, err
		}

		deps[dep.Name] = dep
	}

	return deps, nil
}
