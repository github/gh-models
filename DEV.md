# Developing

## Prerequisites

The extension requires the [`gh` CLI](https://cli.github.com/) to be installed and added to the `PATH`. Users must also
authenticate via `gh auth` before using the extension.

For development, we use [Go](https://golang.org/) with a minimum version of 1.22.

```shell
$ go version
go version go1.22.x <arch>
```

## Building

To build the project, run `script/build`. After building, you can run the binary locally, for example:
`./gh-models list`.

## Testing

To run lint tests, unit tests, and other Go-related checks before submitting a pull request, use:

```shell
make check
```

We also provide separate scripts for specific tasks, where `check` runs them all:

```shell
make test
make fmt  # for auto-formatting
make vet  # to find suspicious constructs
make tidy # to keep dependencies up-to-date
```

## Releasing

When upgrading or installing the extension using `gh extension upgrade github/gh-models` or
`gh extension install github/gh-models`, the latest release will be pulled, not the latest commit. Therefore, all
changes require a new release:

```shell
git tag v0.0.x main
git push origin tag v0.0.x
```

This process triggers the `release` action, which runs the production build.
