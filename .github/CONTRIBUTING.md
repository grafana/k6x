# Contributing to k6x

Thank you for your interest in contributing to k6x!

Before you begin, make sure to familiarize yourself with the [Code of Conduct](CODE_OF_CONDUCT.md). If you've previously contributed to other open source project, you may recognize it as the classic [Contributor Covenant](https://contributor-covenant.org/).

If you want to chat with the team or the community, you can [join our community forums](https://community.grafana.com/c/grafana-k6/).

## Filing issues

Don't be afraid to file issues! Nobody can fix a bug we don't know exists, or add a feature we didn't think of.

The worst that can happen is that someone closes it and points you in the right direction.

That said, "how do I..."-type questions are often more suited for community forums.

## Contributing code

If you'd like to contribute code to k6x, this is the basic procedure. Make sure to follow the [code style](#code-style) described below.

1. Find an issue you'd like to fix. If there is none already, or you'd like to add a feature, please open one, and we can talk about how to do it.  Out of respect for your time, please start a discussion regarding any bigger contributions either in a GitHub Issue, in the community forums **before** you get started on the implementation.
  
   Remember, there's more to software development than code; if it's not properly planned, stuff gets messy real fast.

2. Create a fork and open a feature branch - `feature/my-cool-feature` is the classic way to name these, but it really doesn't matter.

3. Create a pull request!

4. We will discuss implementation details until everyone is happy, then a maintainer will merge it.

### Code style

As you'd expect, please adhere to good ol' `gofmt` (there are plugins for most editors that can autocorrect this), but also `gofmt -s` (code simplification), and don't leave unused functions laying around.

Continuous integration will catch all of this if you don't, and it's fine to just fix linter complaints with another commit, but you can also run the linter yourself:

```bash
golangci-lint run
```

Comments in the source should wrap at 100 characters, but there's no maximum length or need to be brief here - please include anything one might need to know in order to understand the code, that you could reasonably expect any reader to not already know (you probably don't need to explain what a goroutine is).

### Commit format

We don't have any explicit rules about commit message formatting, but try to write something that could be included as-is in a changelog.

If your commit closes an issue, please [close it with your commit message](https://help.github.com/articles/closing-issues-via-commit-messages/), for example:

```text
Added this really rad feature

Closes #420
```

### Language and text formatting

Any human-readable text you add must be non-gendered and should be fairly concise without devolving into grammatical horrors, dropped words, and shorthands. This isn't Twitter, you don't have a character cap, but don't write a novel where a single sentence would suffice.

If you're writing a longer block of text to a terminal, wrap it at 80 characters - this ensures it will display properly at the de facto default terminal size of 80x25.


### Development setup

To get a basic development environment for Go and k6x up and running, first make sure you have **[Git](https://git-scm.com/downloads)** and **[Go](https://golang.org/doc/install)** (see our [go.mod](https://github.com/szkiba/k6x/blob/master/go.mod) for minimum required version) installed and working properly.

We recommend using the Git command-line interface to download the source code for the k6x:

* Open a terminal and run `git clone https://github.com/grafana/k6x.git`. This command downloads k6x's sources to a new `k6x` directory in your current directory.
* Open the `k6x` directory in your favorite code editor.

For alternative ways of cloning the k6x repository, please refer to [GitHub's cloning a repository](https://docs.github.com/en/github/creating-cloning-and-archiving-repositories/cloning-a-repository) documentation.

### Tasks

This section contains a description of the tasks performed during development. Commands must be issued in the `k6x` base directory. If you have the [xc](https://github.com/joerdav/xc) command-line tool, individual tasks can be executed simply by using the `xc task-name` command in the `k6x` base directory.

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
go build -ldflags="-w -s" -o build/k6x .
```

#### snapshot

Creating an executable binary with a snapshot version.

The goreleaser command-line tool is used during the release process. During development, it is advisable to create binaries with the same tool from time to time.

```
goreleaser build --snapshot --clean --single-target -o build/k6x
```

#### clean

Delete the build directory.

```
rm -rf build
```
