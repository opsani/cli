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

Once initialized, you can work with the app using the subcommands of `opsani app`.

Help is available via `opsani --help`.

### Persistent & Ad-hoc Invocations

The Opsani CLI is designed to be a flexible utility that is useful in day to day
development & administration activities and suitable for integration into automated
workflows. The CLI can maintain a persistent configuration in ~/.opsani or can be
configured via environment variables and CLI arguments for automation tasks.

Certain functionality such as the profile and servo registries are dependent on
persistent storage and will emit an error if they are unavailable due to a configuration
file not found condition.

### Profiles

To support users who work across a number of Opsani applications, the CLI supports
named profiles. When the client is initialized, the first profile is named "default"
and is auto-selected when no profile argument is supplied. Profiles can be managed via
the `opsani profile` subcommands.

## Documentation

The primary source of documentation at this stage is this README and the CLI help text.

## Installation

Opsani CLI is distributed in several forms to support easy installation.

### macOS (Homebrew)

Builds for macOS systems can be installed via Homebrew:

```console
$ brew tap opsani/tap
$ brew install opsani-cli
```

### Windows (scoop)

Builds for Windows systems are available via scoop:

```console
$ scoop bucket add github-gh https://github.com/opsani/scoop-bucket.git
$ scoop install opsani-cli
```

### Linux

Binary packages for Linux are available in RPM and DEB package formats.
Download the appropriate package for your distribution of the [latest release](https://github.com/opsani/cli/releases/latest)
and then install as detailed below.

#### Debian/Ubuntu Linux

`$ sudo apt install ./opsani-cli_*_linux_amd64.deb`

#### Fedora Linux

`$ sudo dnf install opsani-cli_*_linux_amd64.rpm`

#### Centos Linux

`$ sudo yum localinstall opsani-cli_*_linux_amd64.rpm`

#### openSUSE/SUSE Linux

`$ sudo zypper in opsani-cli_*_linux_amd64.rpm`

### Binary Releases

Versioned binaries for all platforms are distributed via GitHub: https://github.com/opsani/cli/releases.

Download and install the appropriate build for your platform

### From Source

The `Makefile` is configured with a `make install` target that will build and
install the CLI into `/usr/local/bin/opsani` on Unixy platforms.

## Running via Docker

Containerized releases of Opsani CLI are pushed to the Opsani organization on Docker Hub.
To retrieve the latest container image:

```console
$ docker pull opsani/cli:latest
or
$ make image
```

Once the image is pulled, configuration must be provided to enable the CLI to function. There are
a couple of options for supplying the configuration:

1. **CLI Arguments** - Arguments can be directly supplied to the CLI via the Docker invocation: 
`docker run -it --rm --name opsani-cli opsani/cli:latest --app example.com/app --token 123456`
2. **Environment Variables** - Export `OPSANI_APP` and `OPSANI_TOKEN` into your shell and pass them
through to the container: `docker run -it --rm --name opsani-cli -e OPSANI_APP=$OPSANI_APP -e OPSANI_TOKEN=$OPSANI_TOKEN opsani/cli:latest`
3. **Config Volume** - Existing configuration files in your home directory can be shared with the container through a
volume mount: `docker run -it --rm --name opsani-cli -v ~/.opsani:/root/.opsani opsani/cli:latest`

Commands can then be appended to the end of the command just as when executing in a local shell.

Note that running under a container for day to day work is typically unnecessary and inconvenient
as Opsani CLI is distributed as a single statically linked binary.

### Building a local image

This repository includes a Dockerfile for building and running Opsani CLI in a
container. To build the image from the Dockerfile:

```console
$ docker build . -t opsani/cli:latest
```

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
