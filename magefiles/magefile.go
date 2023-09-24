// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

//go:build mage

package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/princjef/mageutil/bintool"
)

// download required build tools
func Tools() error {
	return tools()
}

// run the golangci-lint linter
func Lint() error {
	return lint()
}

// Build build the binary
func Build() error {
	return build()
}

// Generate generate go sources and assets
func Generate() error {
	return generate()
}

// run tests
func Test() error {
	return test()
}

// show HTML coverage report
func Cover() error {
	return cover()
}

// remove temporary build files
func Clean() error {
	return clean()
}

// lint, test, build
func All() error {
	if err := Lint(); err != nil {
		return err
	}

	if err := Test(); err != nil {
		return err
	}

	return Build()
}

// update license headers
func License() error {
	return license()
}

// ---------------------------------------

var (
	bindir  string
	workdir string
	module  string
)

func init() {
	cwd, err := os.Getwd()
	must(err)

	bindir = filepath.Join(cwd, ".bin")
	workdir = filepath.Join(cwd, "build")

	os.MkdirAll(bindir, 0o755)
	os.MkdirAll(workdir, 0o755)

	path := fmt.Sprintf("%s%c%s", bindir, os.PathListSeparator, os.Getenv("PATH"))
	os.Setenv("PATH", path)

	mod, err := os.ReadFile("go.mod")
	must(err)

	module = string(mod[7:bytes.IndexRune(mod, '\n')])
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func exists(filename string) bool {
	if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
		return false
	}

	return true
}

func goinstall(target string) error {
	return sh.RunWith(map[string]string{"GOBIN": bindir}, "go", "install", target)
}

func findPip() (string, bool) {
	if _, err := exec.LookPath("pipx"); err == nil {
		return "pipx", true
	}

	if _, err := exec.LookPath("pip"); err == nil {
		return "pip", true
	}

	return "", false
}

func hasReuse() bool {
	if _, err := exec.LookPath("reuse"); err != nil {
		return false
	}

	return true
}

// tools downloads k6 golangci-lint configuration, golangci-lint and xk6 binary.
func tools() error {
	resp, err := http.Get("https://raw.githubusercontent.com/grafana/k6/master/.golangci.yml")
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to download linter configuration")
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = os.WriteFile(".golangci.yml", content, 0o644)
	if err != nil {
		return err
	}

	if !bytes.HasPrefix(content, []byte("# v")) {
		return errors.New("missing version comment")
	}

	version := strings.TrimSpace(string(content[3:bytes.IndexRune(content, '\n')]))

	linter, err := bintool.New(
		"golangci-lint{{.BinExt}}",
		version,
		"https://github.com/golangci/golangci-lint/releases/download/v{{.Version}}/golangci-lint-{{.Version}}-{{.GOOS}}-{{.GOARCH}}{{.ArchiveExt}}",
		bintool.WithFolder(bindir),
	)
	if err != nil {
		return err
	}

	if linter.IsInstalled() {
		return nil
	}

	if pip, ok := findPip(); ok && !hasReuse() {
		if err := sh.Run(pip, "install", "reuse"); err != nil {
			return err
		}
	}

	return linter.Ensure()
}

func lint() error {
	mg.Deps(tools)

	_, err := sh.Exec(nil, os.Stdout, os.Stderr, "golangci-lint", "run")
	if err != nil {
		return err
	}

	if hasReuse() {
		_, err := sh.Exec(nil, os.Stdout, os.Stderr, "reuse", "lint", "-q")

		return err
	}

	return nil
}

func generate() error {
	mg.Deps(tools)

	return license()
}

func coverprofile() string {
	return filepath.Join(workdir, "coverage.txt")
}

func test() error {
	maxproc := "4"

	if runtime.GOOS == "windows" {
		maxproc = "1"
	}

	env := map[string]string{
		"GOMAXPROCS": maxproc,
	}

	_, err := sh.Exec(
		env,
		os.Stdout,
		os.Stderr,
		"go",
		"test",
		"-count",
		"1",
		"-p",
		maxproc,
		"-race",
		"-coverprofile="+coverprofile(),
		"./...",
	)

	return err
}

func build() error {
	_, err := sh.Exec(
		nil,
		os.Stdout,
		os.Stderr,
		"go",
		"build",
		"-o",
		"k6x",
		"--ldflags",
		"-s -w",
		".",
	)

	return err

}

func cover() error {
	mg.Deps(test)
	_, err := sh.Exec(nil, os.Stdout, os.Stderr, "go", "tool", "cover", "-html="+coverprofile())
	return err
}

func clean() error {
	sh.Rm("build")
	sh.Rm("dist")
	sh.Rm(".bin")

	return nil
}

func license() error {
	mg.Deps(tools)

	if !hasReuse() {
		fmt.Println("reuse tool missing, you should update license information manually")
		return nil
	}

	return sh.Run(
		"reuse", "annotate",
		"--copyright", "Iván SZKIBA",
		"--merge-copyrights",
		"--license", "AGPL-3.0-only",
		"--skip-unrecognised", "--recursive", ".",
	)
}
