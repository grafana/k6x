// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

//nolint:revive
package resolver

import (
	"net/url"
	"strings"

	"github.com/szkiba/k6x/internal/dependency"
)

type extensionRegistry struct {
	Extensions []registeredExtension `json:"extensions,omitempty"`
}

type registeredExtension struct {
	Name string   `json:"name,omitempty"`
	URL  string   `json:"url,omitempty"`
	Type []string `json:"type,omitempty"`
}

func (reg *extensionRegistry) toModules() dependency.Modules {
	mods := make(dependency.Modules)

	add := func(path, name string) {
		ing := &dependency.Module{
			Path:     path,
			Artifact: &dependency.Artifact{Name: name, Version: nil},
		}
		mods[ing.Name] = ing
	}

	for _, regExt := range reg.Extensions {
		loc, err := url.Parse(regExt.URL)
		if err != nil {
			continue
		}

		path := loc.Host + loc.Path

		for _, typ := range regExt.Type {
			if typ == "Output" {
				add(path, regExt.Name)
				add(path, strings.TrimPrefix(regExt.Name, "xk6-output-"))
				add(path, strings.TrimPrefix(regExt.Name, "xk6-"))
			}

			if typ == "JavaScript" {
				add(path, "k6/x/"+strings.TrimPrefix(regExt.Name, "xk6-"))

				if idx := strings.LastIndex(regExt.Name, "-"); idx >= 0 && idx < len(regExt.Name) {
					add(path, "k6/x/"+regExt.Name[idx+1:])
				}
			}
		}
	}

	return mods
}

func (reg *extensionRegistry) toUniqueModules() dependency.Modules {
	mods := make(dependency.Modules)

	add := func(path, name string) {
		ing := &dependency.Module{
			Path:     path,
			Artifact: &dependency.Artifact{Name: name, Version: nil},
		}
		mods[ing.Name] = ing
	}

	for _, regExt := range reg.Extensions {
		loc, err := url.Parse(regExt.URL)
		if err != nil {
			continue
		}

		path := loc.Host + loc.Path

		for _, typ := range regExt.Type {
			if typ == "Output" {
				add(path, strings.TrimPrefix(regExt.Name, "xk6-"))
			}

			if typ == "JavaScript" {
				if idx := strings.LastIndex(regExt.Name, "-"); idx >= 0 && idx < len(regExt.Name) {
					add(path, "k6/x/"+regExt.Name[idx+1:])
				} else {
					add(path, "k6/x/"+strings.TrimPrefix(regExt.Name, "xk6-"))
				}
			}
		}
	}

	return mods
}
