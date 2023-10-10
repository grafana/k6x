// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

//nolint:revive
package builder

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

var ErrInvalidPlatform = errors.New("invalid platform")

type Platform struct {
	OS   string
	Arch string
}

func RuntimePlatform() *Platform {
	return &Platform{OS: runtime.GOOS, Arch: runtime.GOARCH}
}

func NewPlatform(os, arch string) *Platform {
	return &Platform{OS: os, Arch: arch}
}

func ParsePlatform(str string) (*Platform, error) {
	idx := strings.IndexRune(str, '/')
	if idx <= 0 || idx == len(str)-1 {
		return nil, fmt.Errorf("%w: %s", ErrInvalidPlatform, str)
	}

	return NewPlatform(str[:idx], str[idx+1:]), nil
}

func (p *Platform) String() string {
	return p.OS + "/" + p.Arch
}

func (p *Platform) Supported() bool {
	for _, plat := range supported {
		if plat.OS == p.OS && plat.Arch == p.Arch {
			return true
		}
	}

	return false
}

func SupportedPlatforms() []*Platform {
	return append([]*Platform{}, supported...)
}

var supported = []*Platform{ //nolint:gochecknoglobals
	{OS: "linux", Arch: "amd64"},
	{OS: "linux", Arch: "arm64"},
	{OS: "windows", Arch: "amd64"},
	{OS: "windows", Arch: "arm64"},
	{OS: "darwin", Arch: "amd64"},
	{OS: "darwin", Arch: "arm64"},
}
