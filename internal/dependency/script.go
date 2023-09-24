// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

//nolint:revive,gochecknoglobals
package dependency

import (
	"errors"
	"fmt"
	"io"
	"path"
	"regexp"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/spf13/afero"
)

var (
	reRequire = regexp.MustCompile(
		`require\("((?P<local>\.[^"]+)|((?P<extension>k6/x/[^/^"]+)(/[^"]+)?)|([^"]+))"\)`,
	)
	idxRequireLocal     = reRequire.SubexpIndex("local")
	idxRequireExtension = reRequire.SubexpIndex("extension")

	reUseK6 = regexp.MustCompile(
		`"use k6(( with (?P<extName>(k6/x/)?[0-9a-zA-Z_-]+)( +(?P<extConstraints>[vxX*|,&\^0-9.+-><=, ~]+))?)|(( +(?P<k6Constraints>[vxX*|,&\^0-9.+-><=, ~]+)?)))"`, //nolint:lll
	)
	idxExtName        = reUseK6.SubexpIndex("extName")
	idxExtConstraints = reUseK6.SubexpIndex("extConstraints")
	idxK6Constraints  = reUseK6.SubexpIndex("k6Constraints")

	ErrScript = errors.New("script error")
)

func FromScript(filename string, fs afero.Fs, extra Dependencies) (Dependencies, error) {
	visited := make(map[string]struct{})
	found := make(Dependencies)

	err := findDependencies(filename, fs, found, visited)
	if err != nil {
		return nil, err
	}

	found.ensure(extra)

	return found, nil
}

func findDependencies(
	filename string,
	fs afero.Fs,
	found Dependencies,
	visited map[string]struct{},
) error {
	dir := path.Dir(filename)

	if _, done := visited[filename]; done {
		return nil
	}

	visited[filename] = struct{}{}

	file, err := fs.Open(filename)
	if err != nil {
		return scriptError(err, filename)
	}

	defer deferredClose(file, &err)

	src, err := loadScript(filename, file)
	if err != nil {
		return scriptError(err, filename)
	}

	for _, match := range reRequire.FindAllSubmatch(src, -1) {
		if extension := string(match[idxRequireExtension]); len(extension) != 0 {
			_ = found.update(&Dependency{Name: extension}) // no chance for conflicting
		}

		if local := string(match[idxRequireLocal]); len(local) != 0 {
			err := findDependencies(path.Join(dir, local), fs, found, visited)
			if err != nil {
				return scriptError(err, filename)
			}
		}
	}

	return processUseDirectives(filename, src, found)
}

func processUseDirectives(filename string, src []byte, found Dependencies) error {
	for _, match := range reUseK6.FindAllSubmatch(src, -1) {
		var dep *Dependency
		var err error

		if constraints := string(match[idxK6Constraints]); len(constraints) != 0 {
			dep, err = New(k6, constraints)
			if err != nil {
				return scriptError(err, filename)
			}
		}

		if extension := string(match[idxExtName]); len(extension) != 0 {
			constraints := string(match[idxExtConstraints])

			dep, err = New(extension, constraints)
			if err != nil {
				return scriptError(err, filename)
			}
		}

		if dep != nil {
			if err := found.update(dep); err != nil {
				return scriptError(err, filename)
			}
		}
	}

	return nil
}

func loadScript(filename string, reader io.Reader) ([]byte, error) {
	raw, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	result := api.Transform(string(raw), api.TransformOptions{ //nolint:exhaustruct
		LogLevel:   api.LogLevelSilent,
		Target:     api.DefaultTarget,
		Platform:   api.PlatformDefault,
		Format:     api.FormatCommonJS,
		Sourcefile: filename,
	})

	if len(result.Errors) > 0 {
		msg := result.Errors[0]
		return nil, fmt.Errorf(
			"%s:%d:%d: %w: %s",
			filename,
			msg.Location.Line,
			msg.Location.Column,
			ErrScript,
			msg.Text,
		)
	}

	return result.Code, nil
}

func scriptError(err error, filename string) error {
	return fmt.Errorf("%s: %w: %s", filename, ErrScript, err.Error())
}
