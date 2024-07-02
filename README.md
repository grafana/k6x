[![GitHub Release](https://img.shields.io/github/v/release/grafana/k6x)](https://github.com/grafana/k6x/releases/)
[![Go Report Card](https://goreportcard.com/badge/github.com/grafana/k6x)](https://goreportcard.com/report/github.com/grafana/k6x)
[![GitHub Actions](https://github.com/grafana/k6x/actions/workflows/test.yml/badge.svg)](https://github.com/grafana/k6x/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/grafana/k6x/graph/badge.svg?token=nQA0QAF85R)](https://codecov.io/gh/grafana/k6x)
![GitHub Downloads](https://img.shields.io/github/downloads/grafana/k6x/total)

<h1 name="title">k6x</h1>

> [!Important]
> The k6x is under refactor. This documentation is about the refactored code. Previous documentation is marked with the [before-grafana](https://github.com/grafana/k6x/tree/before-grafana) git tag. The last release before the refactor is [v0.4.0](https://github.com/grafana/k6x/releases/tag/v0.4.0).

**Run k6 with extensions**

k6x is a [k6](https://k6.io) launcher that automatically provides k6 with the [extensions](https://k6.io/docs/extensions/) used by the test. In order to do this, it analyzes the script arguments of the `run` and `archive` subcommands, detects the extensions to be used and their version constraints.

## Install

Precompiled binaries can be downloaded and installed from the [Releases](https://github.com/grafana/k6x/releases) page.

If you have a go development environment, the installation can also be done with the following command:

```
go install github.com/grafana/k6x@latest
```

The launcher acts as a drop-in replacement for the k6 command. For more convenient use, it is advisable to create an alias or shell script called k6 for the launcher. The alias can be used in exactly the same way as the k6 command, with the difference that it generates the real k6 on the fly based on the extensions you want to use.

## Usage

<!-- #region cli -->
## k6x

Lanch k6 with extensions

### Synopsis

Launch k6 with a seamless extension user experience.

The launcher acts as a drop-in replacement for the `k6` command. For more convenient use, it is advisable to create an alias or shell script called `k6` for the launcher. The alias can be used in exactly the same way as the `k6` command, with the difference that it generates the real `k6` on the fly based on the extensions you want to use.

The launcher will always run the k6 test script with the appropriate k6 binary, which contains the extensions used by the script. In order to do this, it analyzes the script arguments of the "run" and "archive" subcommands, detects the extensions to be used and their version constraints.

Any k6 command can be used. Use the `help` command to list the available k6 commands.


```
k6x [flags] [command]
```

### Flags

```
      --build-service-url string       URL of the k6 build service to be used
      --extension-catalog-url string   URL of the k6 extension catalog to be used
  -h, --help                           help for k6
      --no-color                       disable colored output
  -q, --quiet                          disable progress updates
      --usage                          print launcher usage
  -v, --verbose                        enable verbose logging
```

<!-- #endregion cli -->

## Contributing

If you want to contribute, please read the [CONTRIBUTING.md](.github/CONTRIBUTING.md) file.

### Tasks

This section contains a description of the tasks performed during development. Commands must be issued in the repository base directory. If you have the [xc](https://github.com/joerdav/xc) command-line tool, individual tasks can be executed simply by using the `xc task-name` command in the repository base directory.

<details><summary>Click to expand</summary>

#### readme

Update documentation in README.md.

```
go run ./tools/gendoc README.md
```

#### lint

Run the static analyzer.

We make use of the [golangci-lint](https://github.com/golangci/golangci-lint) tool to lint the code in CI. The actual version you can find in our [`.golangci.yml`](https://github.com/grafana/k6x/blob/master/.golangci.yml#L1).

```
golangci-lint run
```

#### test

Run the tests.

To exercise the entire test suite, please run the following command


```
go test -count 1 -race -coverprofile=build/coverage.txt ./...
```

#### coverage

View the test coverage report.

```
go tool cover -html=build/coverage.txt
```

#### build

Build the executable binary.

This is the easiest way to create an executable binary (although the release process uses the goreleaser tool to create release versions).

```
go build -ldflags="-w -s" -o k6x .
```

#### snapshot

Creating an executable binary with a snapshot version.

The goreleaser command-line tool is used during the release process. During development, it is advisable to create binaries with the same tool from time to time.

```
goreleaser build --snapshot --clean --single-target -o k6x
```

#### docker

Building a Docker image. Before building the image, it is advisable to perform a snapshot build using goreleaser. To build the image, it is advisable to use the same `Docker.goreleaser` file that `goreleaser` uses during release.

Requires: snapshot

```
docker build -t grafana/k6x -f Dockerfile.goreleaser .
```

#### clean

Delete the build directory.

```
rm -rf build
```

</details>
