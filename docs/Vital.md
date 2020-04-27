# Opsani Vital

This document details the integration between the Opsani CLI application and the
Opsani Vital product offering.

## Installation & Activation experience

1. Start the experience via Docker: `make run`
1. Open the sign-up experience at http://localhost:8888/
    1. Input your name
    2. Input your work e-mail
    3. Input your app name (Optional, auto-generated if blank)
5. Upon completion, send an email that includes:
    1. The name, e-mail, and app name
    1. A one-time use token (i.e. `2ada27a6cf09`)
    2. Instructions for installing the CLI automatically: 
        `curl http://localhost:8888/install.sh/2ada27a6cf09 | sh`
    3. A link to a help page for installing manually:
        `http://localhost:8888/install`
    4. Instructions for activating if you already have Opsani CLI: 
        Run `opsani init 2ada27a6cf09` to get started
6. When executed, the script detects your host OS and fetches the appropriate
   CLI build and installs it (use the Rust installer script as a reference)
    1. Gracefully handle fall-back if we can't auto-detect and point user at the
       manual install page
7. When the user runs `opsani init 2ada27a6cf09`
    1. Configuration is auto-discovered via the one-time use token
    2. A config file is written to `~/opsani/config.yaml`
9. Upon success, display a message about starting with an interactive tutorial
   via `opsani tutorial` or configuring your real app with `opsani vital`.

## Implementing the Vital experience

These steps will be automated into the Makefile as `make vital`.

1. Create a Dockerfile in vital/ with Nginx, a simple HTTP service, and an email
   library.
    1. Configure the index to return the CLI installer shell script (see below)
    2. Add index.html into the doc-root for displaying the sign-up page
    3. Set-up a simple dynamic HTTP handler on `/signup` that accepts a POST of
       `name`, `email`, and `app_name`
    4. Send an email to the user as described above
2. Create a .env file with email credentials
3. Generate a snapshot build of CLI for all platforms using go-releaser
4. Copy snapshot into vital/builds/
5. Map 8888 as a public port
1. Create a Makefile for the experience
    1. `make build` builds the experience via Docker and tags
       `opsani/vital:latest`
    2. `make run` runs the container and launches the experience in your browser
