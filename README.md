# Opsani CLI

Opsani CLI is cloud optimization in your terminal. It brings a suite of tools for 
configuring & deploying Servos, managing optimization runs, and interacting with the optimization engine
to your command line.

## Status

Opsani CLI is an early stage project in active development.

We need your feedback to determine how best to evolve the tool.

If you run into any bugs or think of a feature you'd like to see, please file a ticket on GitHub.

## Usage

The CLI currently supports controling optimization lifecycle, managing configuration, and building
Servo artifacts for connecting new apps to Opsani for optimization.

To perform any meaningful work, you must first initialize the client via `opsani init` and supply
details about your app and API token.

Once initialized, you can work with the app using the subcommands of `opsani app`.

The `opsani discover` commands is an evolving system for auto-discovering the environment
to facilitate rapid integration with the optimization engine. Discovery connects to Docker,
Kubernetes, and Prometheus and can execute over an SSH channel via `opsani discover --host ssh://user@host`.

Help is available via `opsani --help`.

## Documentation

Docs are forthcoming. Utilize the CLI help and read the code for now.

## Installation

Versioned releases are coming soon. Please build from source in the interim.

`make install` is available on Unixy platforms and will drop the binary into /usr/local/bin. 

## Development

Opsani CLI is implemented in Golang and utilizes a handful of well established libraries including
Cobra, Viper, and Resty. Anyone reasonably well versed in Go should have no trouble following along.

There is a Makefile for running typical tasks but `go run .` is a great way to poke around.

The code is in pretty decent shape but factoring is very much a moving target.

## License

Opsani CLI is released under the Apache 2.0 license. See LICENSE
