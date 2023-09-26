// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

//nolint:revive
package builder

type Engine int

const (
	Native Engine = iota
	Docker
)

//go:generate enumer -json -text -values -type Engine
