# Opsani CLI

This repository contains a minimal experiment for implementing an Opsani CLI in Golang.

The primary experiment here is implementing a pure Golang CLI orchestrator which wraps around
modules implemented as Docker containers, facilitating a hybrid implementation path where some
code is native Golang and some is written in arbitrary languages executed within a container.

## Usage

The command expects a locally built image of the Intelligent Manifest Builder (IMB) tagged into
your local Docker environment as `imb`:

1. Clone the Dockerized build of IMB: `git clone git@github.com:opsani/intelligent-manifest-builder.git cli-experiments`
2. Build the Docker tag: `docker build -t opsani/intelligent-manifest-builder .`
3. Run the CLI: `go run . --help`
4. Run local discovery and generate Servo assets: `go run . discover`
5. Run remote discovery over  SSH and generate Servo assets: `go run main.go discover --host ssh://localhost`

Help is available via `go run . --help`.
