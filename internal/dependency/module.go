// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

//nolint:revive
package dependency

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
)

const (
	latestTag = "latest"

	versionTagPrefix = "v"
)

type Module struct {
	Name    string          `json:"name,omitempty"`
	Version *semver.Version `json:"version,omitempty"`
	Path    string          `json:"path,omitempty"`
}

func NewModule(name, version, path string) (*Module, error) {
	var err error
	mod := new(Module)

	mod.Name = name
	mod.Path = path

	if len(version) != 0 {
		mod.Version, err = semver.NewVersion(version)
		if err != nil {
			return nil, err
		}
	}

	return mod, nil
}

func (mod *Module) Tag() string {
	if mod.Version == nil {
		return latestTag
	}

	return versionTagPrefix + mod.Version.String()
}

func (mod *Module) String() string {
	var buff strings.Builder

	buff.WriteString(mod.Name)
	buff.WriteRune(' ')
	buff.WriteString(mod.Path)
	buff.WriteRune(' ')
	buff.WriteString(mod.Tag())

	return buff.String()
}

func (mod *Module) Ref() string {
	var buff strings.Builder

	buff.WriteString(mod.Name)
	buff.WriteRune('@')
	buff.WriteString(mod.Tag())

	return buff.String()
}

type Modules map[string]*Module

func (mods Modules) Resolves(deps Dependencies) bool {
	for _, dep := range deps {
		mod, ok := mods[dep.Name]
		if !ok {
			return false
		}

		if !dep.Check(mod.Version) {
			return false
		}
	}

	return true
}

func (mods Modules) K6() (*Module, bool) {
	mod, ok := mods[k6]

	return mod, ok
}

func (mods Modules) Extensions() []*Module {
	exts := make([]*Module, 0, len(mods))

	for name, dep := range mods {
		if name != k6 {
			exts = append(exts, dep)
		}
	}

	sort.Slice(exts, func(i, j int) bool {
		return exts[i].Name < exts[j].Name
	})

	return exts
}

func (mods Modules) Sorted() []*Module {
	all := make([]*Module, 0, len(mods))

	for _, dep := range mods {
		all = append(all, dep)
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

func (mods Modules) String() string {
	var buff strings.Builder

	for _, mod := range mods.Sorted() {
		buff.WriteString(mod.String())
		buff.WriteRune('\n')
	}

	return buff.String()
}

func (mods Modules) Ref() string {
	var buff strings.Builder

	for _, mod := range mods.Sorted() {
		if buff.Len() != 0 {
			buff.WriteRune(',')
		}

		buff.WriteString(mod.Ref())
	}

	return buff.String()
}

func (mods Modules) Filter(deps Dependencies) Modules {
	found := make(Modules)

	for _, dep := range deps {
		if ing, ok := mods[dep.Name]; ok && dep.Check(ing.Version) {
			found[dep.Name] = ing
		}
	}

	return found
}

func (mods Modules) MarshalJSON() ([]byte, error) {
	dict := make(map[string]string, len(mods))

	for _, mod := range mods {
		var buff strings.Builder

		buff.WriteString(mod.Path)
		buff.WriteRune('@')
		buff.WriteString(mod.Tag())

		dict[mod.Name] = buff.String()
	}

	return json.Marshal(dict)
}
