module github.com/opsani/cli

go 1.14

require (
	github.com/AlecAivazis/survey/v2 v2.0.7
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/Microsoft/hcsshim v0.8.6 // indirect
	github.com/Netflix/go-expect v0.0.0-20200312175327-da48e75238e2
	github.com/alecthomas/assert v0.0.0-20170929043011-405dbfeb8e38
	github.com/alecthomas/colour v0.1.0 // indirect
	github.com/alecthomas/repr v0.0.0-20200325044227-4184120f674c // indirect
	github.com/charmbracelet/glamour v0.1.0
	github.com/containerd/containerd v1.3.3 // indirect
	github.com/containerd/continuity v0.0.0-20200228182428-0f16d7a0959c // indirect
	github.com/creack/pty v1.1.9 // indirect
	github.com/docker/cli v0.0.0-00010101000000-000000000000
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.13.1
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/dustin/go-humanize v1.0.0
	github.com/fatih/color v1.9.0
	github.com/go-git/go-git v4.7.0+incompatible
	github.com/go-git/go-git/v5 v5.0.0
	github.com/go-resty/resty/v2 v2.2.0
	github.com/goccy/go-yaml v1.4.3
	github.com/golang/protobuf v1.4.0 // indirect
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/googleapis/gnostic v0.4.0 // indirect
	github.com/gorilla/mux v1.7.4 // indirect
	github.com/hinshun/vt10x v0.0.0-20180616224451-1954e6464174
	github.com/hokaccha/go-prettyjson v0.0.0-20190818114111-108c894c2c0e
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/kr/pretty v0.2.0 // indirect
	github.com/kr/pty v1.1.8
	github.com/mattn/go-colorable v0.1.4
	github.com/mattn/go-isatty v0.0.12
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.2.2 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/olekukonko/tablewriter v0.0.4
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/opencontainers/runc v0.1.1 // indirect
	github.com/pelletier/go-toml v1.7.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/common v0.4.0
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/sirupsen/logrus v1.5.0 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.6.3
	github.com/stretchr/testify v1.5.1
	github.com/tidwall/gjson v1.6.0
	github.com/tidwall/sjson v1.1.1
	golang.org/x/crypto v0.0.0-20200302210943-78000ba7a073
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e // indirect
	golang.org/x/sys v0.0.0-20200420163511-1957bb5e6d1f // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	google.golang.org/genproto v0.0.0-20200413115906-b5235f65be36 // indirect
	google.golang.org/grpc v1.28.1 // indirect
	gopkg.in/ini.v1 v1.55.0 // indirect
	gopkg.in/src-d/go-git.v4 v4.13.1 // indirect
	gopkg.in/yaml.v2 v2.2.8
	gotest.tools v2.2.0+incompatible // indirect
	k8s.io/apimachinery v0.18.1
	k8s.io/client-go v0.18.1
	k8s.io/utils v0.0.0-20200411171748-3d5a2fe318e4 // indirect
	sigs.k8s.io/yaml v1.2.0
)

replace github.com/docker/docker => github.com/docker/engine v17.12.0-ce-rc1.0.20200309214505-aa6a9891b09c+incompatible

replace github.com/docker/cli => github.com/docker/cli v0.0.0-20200303215952-eb310fca4956

replace golang.org/x/sys => golang.org/x/sys v0.0.0-20190830141801-acfa387b8d69
