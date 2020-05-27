# Opsani Ignite

Ignite is a demo workflow for Opsani. It quickly configures and deploys
an an application and a servo into a local Kubernetes cluster under minikube.

## Setup (macOS)

1. Install minikube via Homebrew: `$ brew install minikube`
2. Allocate resources to minikube:

```console
$ minikube config set memory 4096
$ minikube config set cpus 4
```

3. Initialize Opsani CLI: `$ opsani init`
4. Run Ignite: `$ opsani ignite`
