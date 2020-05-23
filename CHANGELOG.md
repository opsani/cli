# CHANGELOG

Opsani CLI is an Open Source utility distributed under the terms of the Apache 2.0
license. This changelog catalogs all notable changes made to the project. The format
is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/). Releases are 
versioned in accordance with [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

- Kubernetes discovery and manifest generation.
- Servo assembly and connector management.
- Interactive demo and tutorial experiences.

## [0.1.3] - 2020-05-23
### Added
- Scoop builds for Windows users.

## [0.1.2] - 2020-05-23
### Fixed
- Fixed init failure with missing config file.

## [0.1.1] - 2020-05-23
### Added
- Enabled build and release of RPM and DEB package artifacts.

## [0.1.0] - 2020-05-23

Initial public release.

### Added
- App profile registry for managing apps and tokens.
- Servo registry for managing servo deployments.
- App lifecycle management (start, stop, restart, status).
- App configuration management (get, edit, patch, set).
- Pretty print formatting utilities for JSON and YAML documents.
- Support for executing remote actions over SSH (see servo commands).
- Virtual terminal based testing infrastructure for interactive CLI flows.
- API debugging output and request tracing.

[Unreleased]: https://github.com/opsani/cli/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/opsani/cli/releases/tag/v0.1.0
