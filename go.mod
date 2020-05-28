module github.com/opsani/cli

go 1.14

require (
	github.com/AlecAivazis/survey/v2 v2.0.7
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Netflix/go-expect v0.0.0-20200312175327-da48e75238e2
	github.com/alecthomas/assert v0.0.0-20170929043011-405dbfeb8e38
	github.com/alecthomas/colour v0.1.0 // indirect
	github.com/alecthomas/repr v0.0.0-20200325044227-4184120f674c // indirect
	github.com/briandowns/spinner v1.11.1
	github.com/charmbracelet/glamour v0.1.0
	github.com/docker/docker v1.13.1
	github.com/fatih/color v1.9.0
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/go-resty/resty/v2 v2.3.0
	github.com/goccy/go-yaml v1.4.3
	github.com/google/go-cmp v0.4.0 // indirect
	github.com/hinshun/vt10x v0.0.0-20180616224451-1954e6464174
	github.com/hokaccha/go-prettyjson v0.0.0-20190818114111-108c894c2c0e
	github.com/kr/pty v1.1.8 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/markbates/pkger v0.15.1
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/mattn/go-isatty v0.0.12
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.3.1 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/olekukonko/tablewriter v0.0.4
	github.com/pelletier/go-toml v1.8.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/common v0.4.0
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/sirupsen/logrus v1.6.0 // indirect
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.5.1
	github.com/tidwall/gjson v1.6.0
	github.com/tidwall/sjson v1.1.1
	golang.org/x/crypto v0.0.0-20200510223506-06a226fb4e37
	golang.org/x/net v0.0.0-20200520182314-0ba52f642ac2 // indirect
	golang.org/x/sys v0.0.0-20200523222454-059865788121 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/ini.v1 v1.56.0 // indirect
	gopkg.in/yaml.v2 v2.3.0
	gotest.tools v2.2.0+incompatible // indirect
	sigs.k8s.io/yaml v1.2.0
)

replace github.com/docker/docker => github.com/docker/engine v17.12.0-ce-rc1.0.20200309214505-aa6a9891b09c+incompatible

replace github.com/docker/cli => github.com/docker/cli v0.0.0-20200303215952-eb310fca4956

replace golang.org/x/sys => golang.org/x/sys v0.0.0-20190830141801-acfa387b8d69
