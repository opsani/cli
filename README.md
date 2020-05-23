# Opsani CLI

Opsani CLI is cloud optimization in your terminal. It brings a suite of tools
for configuring & deploying servos, managing optimization runs, and interacting
with the optimization engine to your command line.

## Status

Opsani CLI is an early stage project in active development.

We need your feedback to determine how best to evolve the tool.

If you run into any bugs or think of a feature you'd like to see, please file a
ticket on GitHub.

## Usage

The CLI currently supports controling optimization lifecycle, managing
configuration, and building servo artifacts for connecting new apps to Opsani
for optimization.

To perform any meaningful work, you must first initialize the client via `opsani
init` and supply details about your app and API token.

Once initialized, you can work with the app using the subcommands of `opsani
app`.

Help is available via `opsani --help`.

### Profiles

To support users who work across a number of Opsani applications, the CLI supports
named profiles. When the client is initialized, the first profile is named "default"
and is auto-selected when no profile argument is supplied. Profiles can be managed via
the `opsani profile` subcommands.

## Documentation

The primary source of documentation at this stage is this README and the CLI help text.

## Installation

Opsani CLI is distributed in several forms to support easy installation.

### Binary Releases

Versioned binaries for all platforms are distributed via GitHub: https://github.com/opsani/cli/releases.

To download the latest release for your platform:

```console
$ curl --silent --location "https://github.com/opsani/cli/releases/latest/download/opsani-cli_$(uname -s)_amd64.tar.gz" | tar xz -C /tmp
$ sudo mv /tmp/opsani /usr/local/bin
```

### Homebrew (macOS)

Builds for macOS systems can be installed via Homebrew:

```console
$ brew tap opsani/tap
$ brew install opsani-cli
```

### From Source

The `Makefile` is configured with a `make install` target that will build and
install the CLI into `/usr/local/bin/opsani` on Unixy platforms.

## Development

Opsani CLI is implemented in Golang and utilizes a handful of well established
libraries including Cobra, Viper, and Resty. Anyone reasonably well versed in Go
should have no trouble following along.

There is a Makefile for running typical tasks but `go run .` is a great way to
poke around.

## Testing

Opsani CLI has extensive automated test coverage. Unit tests exist alongside the
code under test and there are integration tests in the `integration` directory.

Testing assertions and mocks are provided by the
[Testify](https://github.com/stretchr/testify) library. Tests are organized
using the `testify.suite` package to provide familiar xUnit testing primitives.

The local test package contains shared testing helpers useful in both unit and
integration test scenarios.

Tests can be run via the Makefile:

* `make test` - Run unit & integration tests.
* `make test_unit` - Run unit tests.
* `make test_integration` - Run integration tests.

### Integration Tests

The integration test harness functions by building the `opsani` binary, copying
it into a temp directory, and then interacting with it as a subprocess and
evaluating stdout, stderr, and exit code status.

API interactions are mocked locally via an `http.httptest` server that runs in
the test executor parent process. The `--base-url` flag is used to direct the
CLI to interact with the test API server rather than the Opsani API.

## License

Opsani CLI is released under the Apache 2.0 license. See LICENSE
