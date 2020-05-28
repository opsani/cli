# Opsani Ignite

Ignite is a demo workflow for Opsani. It quickly configures and deploys
an an application and a servo into a local Kubernetes cluster under minikube.

## Setup (macOS)

1. Install minikube via Homebrew: `$ brew install minikube`
2. Enable elevated permissions for Hyperkit (macOS VM driver):
```console
    $ sudo chown root:wheel ~/.minikube/bin/docker-machine-driver-hyperkit 
    $ sudo chmod u+s ~/.minikube/bin/docker-machine-driver-hyperkit 
```
3. Initialize Opsani CLI: `$ opsani init`
4. Run Ignite: `$ opsani ignite`
