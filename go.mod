module github.com/opsani/cli

go 1.14

require (
	github.com/AlecAivazis/survey/v2 v2.0.7
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/Microsoft/hcsshim v0.8.6 // indirect
	github.com/containerd/containerd v1.3.3 // indirect
	github.com/containerd/continuity v0.0.0-20200228182428-0f16d7a0959c // indirect
	github.com/docker/cli v0.0.0-00010101000000-000000000000
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.13.1
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/go-resty/resty/v2 v2.2.0
	github.com/golang/protobuf v1.4.0 // indirect
	github.com/googleapis/gnostic v0.4.1 // indirect
	github.com/gorilla/mux v1.7.4 // indirect
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/kr/pretty v0.2.0
	github.com/mattn/go-isatty v0.0.9 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.2.2 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.7.0
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/opencontainers/runc v0.1.1 // indirect
	github.com/pelletier/go-toml v1.7.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sirupsen/logrus v1.5.0 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.6.3
	golang.org/x/crypto v0.0.0-20200220183623-bac4c82f6975
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	google.golang.org/genproto v0.0.0-20200413115906-b5235f65be36 // indirect
	google.golang.org/grpc v1.28.1 // indirect
	gopkg.in/ffmt.v1 v1.5.6
	gopkg.in/ini.v1 v1.55.0 // indirect
	gotest.tools v2.2.0+incompatible // indirect
	k8s.io/apimachinery v0.18.1
	k8s.io/client-go v0.18.1
	k8s.io/utils v0.0.0-20200411171748-3d5a2fe318e4 // indirect
)

replace github.com/docker/docker => github.com/docker/engine v17.12.0-ce-rc1.0.20200309214505-aa6a9891b09c+incompatible

replace github.com/docker/cli => github.com/docker/cli v0.0.0-20200303215952-eb310fca4956

replace golang.org/x/sys => golang.org/x/sys v0.0.0-20190830141801-acfa387b8d69
