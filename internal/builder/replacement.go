// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

//nolint:revive
package builder

import (
	"context"
	"encoding/json"
	"strings"
)

type Replacement struct {
	Name string `json:"name,omitempty"`
	Path string `json:"path,omitempty"`
}

func NewReplacement(name, path string) *Replacement {
	return &Replacement{Name: name, Path: path}
}

func (rep *Replacement) String() string {
	var buff strings.Builder

	buff.WriteString(rep.Name)
	buff.WriteRune(' ')
	buff.WriteString(rep.Path)

	return buff.String()
}

type Replacements map[string]*Replacement

func (reps Replacements) String() string {
	var buff strings.Builder

	for _, rep := range reps {
		buff.WriteString(rep.String())
		buff.WriteRune('\n')
	}

	return buff.String()
}

func (reps Replacements) MarshalJSON() ([]byte, error) {
	dict := make(map[string]string, len(reps))

	for _, rep := range reps {
		dict[rep.Name] = rep.Path
	}

	return json.Marshal(dict)
}

type replacementsKey struct{}

func WithReplacements(ctx context.Context, reps Replacements) context.Context {
	return context.WithValue(ctx, replacementsKey{}, reps)
}

func replacementsFromContext(ctx context.Context) Replacements {
	v := ctx.Value(replacementsKey{})
	if v == nil {
		return make(Replacements)
	}

	if reps, ok := v.(Replacements); ok {
		return reps
	}

	return make(Replacements)
}
