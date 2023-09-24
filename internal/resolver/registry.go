// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

//nolint:revive
package resolver

import (
	"net/url"
	"strings"
)

type extensionRegistry struct {
	Extensions []registeredExtension `json:"extensions,omitempty"`
}

type registeredExtension struct {
	Name string   `json:"name,omitempty"`
	URL  string   `json:"url,omitempty"`
	Type []string `json:"type,omitempty"`
}

func (reg *extensionRegistry) toIngredients() Ingredients {
	ings := make(Ingredients)

	add := func(module, name string) {
		ing := &Ingredient{Module: module, Name: name, Version: nil}
		ings[ing.Name] = ing
	}

	for _, regExt := range reg.Extensions {
		loc, err := url.Parse(regExt.URL)
		if err != nil {
			continue
		}

		module := loc.Host + loc.Path

		for _, typ := range regExt.Type {
			if typ == "Output" {
				add(module, regExt.Name)
				add(module, strings.TrimPrefix(regExt.Name, "xk6-output-"))
				add(module, strings.TrimPrefix(regExt.Name, "xk6-"))
			}

			if typ == "JavaScript" {
				add(module, "k6/x/"+strings.TrimPrefix(regExt.Name, "xk6-"))

				if idx := strings.LastIndex(regExt.Name, "-"); idx >= 0 && idx < len(regExt.Name) {
					add(module, "k6/x/"+regExt.Name[idx+1:])
				}
			}
		}
	}

	return ings
}
