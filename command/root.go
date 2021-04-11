// Copyright 2020 Opsani
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package command

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"
	"text/template"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/core"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/briandowns/spinner"
	"github.com/charmbracelet/glamour"
	"github.com/docker/docker/pkg/term"
	"github.com/fatih/color"
	"github.com/mitchellh/go-homedir"
	"github.com/opsani/cli/opsani"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/scheme"

	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"

	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// "k8s.io/client-go/kubernetes"
	// "k8s.io/client-go/tools/clientcmd"
	"github.com/xeonx/timeago"
	k8s_homedir "k8s.io/client-go/util/homedir"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/viper"
	ssh_terminal "golang.org/x/crypto/ssh/terminal"

	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

const manifestTemplate = `---
apiVersion: v1
kind: ConfigMap
metadata:
  name: servo-leviathan-wakes
  labels:
      app.kubernetes.io/name: servo
      app.kubernetes.io/component: core
      servo.opsani.com/optimizer: dev.opsani.com_leviathan-wakes
  annotations:
    servo.opsani.com/optimizer: dev.opsani.com/leviathan-wakes
data:
  optimizer: dev.opsani.com/leviathan-wakes
  log_level: INFO
  servo.yaml: |
    opsani_dev:
      namespace: {{ .Namespace }}
      deployment: {{ .Deployment }}
      container: {{ .Container }}
      service: {{ .Service }}
      cpu:
        min: 250m
        max: '3.0'
      memory:
        min: 128.0MiB
        max: 3.0GiB

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: servo-leviathan-wakes
  labels:
    app.kubernetes.io/name: servo
    app.kubernetes.io/component: core
    servo.opsani.com/optimizer: dev.opsani.com_leviathan-wakes
  annotations:
    servo.opsani.com/optimizer: dev.opsani.com/leviathan-wakes
spec:
  replicas: 1
  revisionHistoryLimit: 2
  strategy:
    type: Recreate
  selector:
    matchLabels:
      app.kubernetes.io/name: servo
      servo.opsani.com/optimizer: dev.opsani.com_leviathan-wakes
  template:
    metadata:
      name: servo-leviathan-wakes
      labels:
        app.kubernetes.io/name: servo
        app.kubernetes.io/component: core
        servo.opsani.com/optimizer: dev.opsani.com_leviathan-wakes
      annotations:
        servo.opsani.com/optimizer: dev.opsani.com/leviathan-wakes
    spec:
      serviceAccountName: servo-leviathan-wakes
      containers:
      - name: servo
        image: opsani/servox:latest
        terminationMessagePolicy: FallbackToLogsOnError
        args:
          - 'check'
          - '--wait=30m'
          - '--delay=10s'
          - '--progressive'
          - '--run'
        env:
        - name: OPSANI_OPTIMIZER
          valueFrom:
            configMapKeyRef:
              name: servo-leviathan-wakes
              key: optimizer
        - name: OPSANI_TOKEN_FILE
          value: /servo/opsani.token
        - name: SERVO_LOG_LEVEL
          valueFrom:
            configMapKeyRef:
              name: servo-leviathan-wakes
              key: log_level
        - name: POD_NAME
          valueFrom:
              fieldRef:
                fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
        volumeMounts:
        - name: servo-token-volume
          mountPath: /servo/opsani.token
          subPath: opsani.token
          readOnly: true
        - name: servo-config-volume
          mountPath: /servo/servo.yaml
          subPath: servo.yaml
          readOnly: true
        resources:
          limits:
            cpu: 250m
            memory: 512Mi
      - name: prometheus
        image: quay.io/prometheus/prometheus:v2.20.1
        args:
          - '--storage.tsdb.retention.time=12h'
          - '--config.file=/etc/prometheus/prometheus.yaml'
        ports:
        - name: webui
          containerPort: 9090
        resources:
          requests:
            cpu: 100m
            memory: 128M
          limits:
            cpu: 500m
            memory: 1G
        volumeMounts:
        - name: prometheus-config-volume
          mountPath: /etc/prometheus
      volumes:
      - name: servo-token-volume
        secret:
          secretName: servo-leviathan-wakes
          items:
          - key: token
            path: opsani.token
      - name: servo-config-volume
        configMap:
          name: servo-leviathan-wakes
          items:
          - key: servo.yaml
            path: servo.yaml
      - name: prometheus-config-volume
        configMap:
          name: servo.prometheus-leviathan-wakes

      # Prefer deployment onto a Node labeled role=servo
      # This ensures physical isolation and network transport if possible
      affinity:
        nodeAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 1
            preference:
              matchExpressions:
              - key: node.opsani.com/role
                operator: In
                values:
                - servo

---
apiVersion: v1
kind: Secret
metadata:
  name: servo-leviathan-wakes
  labels:
    app.kubernetes.io/name: servo
    app.kubernetes.io/component: core
    servo.opsani.com/optimizer: dev.opsani.com_leviathan-wakes
  annotations:
    servo.opsani.com/optimizer: dev.opsani.com/leviathan-wakes
type: Opaque
stringData:
  token: 452361d3-48e0-41df-acec-0fe2be826cb8

---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: servo-leviathan-wakes
  labels:
    app.kubernetes.io/name: servo
    app.kubernetes.io/component: core
    servo.opsani.com/optimizer: dev.opsani.com_leviathan-wakes
  annotations:
    servo.opsani.com/optimizer: dev.opsani.com/leviathan-wakes

---
# Cluster Role for the servo itself
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: servo-leviathan-wakes
  labels:
    app.kubernetes.io/name: servo
    app.kubernetes.io/component: core
    servo.opsani.com/optimizer: dev.opsani.com_leviathan-wakes
  annotations:
    servo.opsani.com/optimizer: dev.opsani.com/leviathan-wakes
rules:
- apiGroups: ["apps", "extensions"]
  resources: ["deployments", "deployments/status", "replicasets"]
  verbs: ["get", "list", "watch", "update", "patch"]
- apiGroups: [""]
  resources: ["pods", "pods/logs", "pods/status", "pods/exec", "pods/portforward", "services"]
  verbs: ["create", "delete", "get", "list", "watch", "update", "patch" ]
- apiGroups: [""]
  resources: ["namespaces"]
  verbs: ["get", "list"]

---
# Cluster Role for the Prometheus sidecar
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: servo.prometheus-leviathan-wakes
  labels:
    app.kubernetes.io/name: prometheus
    app.kubernetes.io/component: metrics
    app.kubernetes.io/part-of: servo
    servo.opsani.com/optimizer: dev.opsani.com_leviathan-wakes
  annotations:
    servo.opsani.com/optimizer: dev.opsani.com/leviathan-wakes
rules:
- apiGroups: [""]
  resources:
  - namespaces
  - nodes
  - nodes/proxy
  - services
  - endpoints
  - pods
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources:
  - configmaps
  - nodes/metrics
  verbs: ["get"]
- nonResourceURLs: ["/metrics"]
  verbs: ["get"]

---
# Bind the Servo Cluster Role to the servo Service Account
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: servo-leviathan-wakes
  labels:
    app.kubernetes.io/name: servo
    app.kubernetes.io/component: core
    servo.opsani.com/optimizer: dev.opsani.com_leviathan-wakes
  annotations:
    servo.opsani.com/optimizer: dev.opsani.com/leviathan-wakes
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: servo-leviathan-wakes
subjects:
- kind: ServiceAccount
  name: servo-leviathan-wakes
  namespace: {{ .Namespace }}

---
# Bind the Prometheus Cluster Role to the servo Service Account
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: servo.prometheus-leviathan-wakes
  labels:
    app.kubernetes.io/name: prometheus
    app.kubernetes.io/component: metrics
    app.kubernetes.io/part-of: servo
    servo.opsani.com/optimizer: dev.opsani.com_leviathan-wakes
  annotations:
    servo.opsani.com/optimizer: dev.opsani.com/leviathan-wakes
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: servo.prometheus-leviathan-wakes
subjects:
- kind: ServiceAccount
  name: servo-leviathan-wakes
  namespace: {{ .Namespace }}

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: servo.prometheus-leviathan-wakes
  labels:
    app.kubernetes.io/name: prometheus
    app.kubernetes.io/component: metrics
    app.kubernetes.io/part-of: servo
    servo.opsani.com/optimizer: dev.opsani.com_leviathan-wakes
  annotations:
    servo.opsani.com/optimizer: dev.opsani.com/leviathan-wakes
data:
  prometheus.yaml: |
    # Opsani Servo Prometheus Sidecar v0.8.0
    # This configuration allows the Opsani Servo to discover and scrape Pods that
    # have been injected with an Envoy proxy sidecar container that emits the metrics
    # necessary for optimization. Scraping by the Prometheus sidecar is enabled by
    # adding the following annotations to the Pod spec of the Deployment under
    # optimization:
    #
    # annotations:
    #   prometheus.opsani.com/scrape: "true" # Opt-in for scraping by the servo
    #   prometheus.opsani.com/scheme: http # Scrape via HTTP by default
    #   prometheus.opsani.com/path: /stats/prometheus # Default Envoy metrics path
    #   prometheus.opsani.com/port: "9901" # Default Envoy metrics port
    #
    # Path and port collisions with the optimization target can be resolved be changing
    # the relevant annotation.

    # Scrape the targets every 5 seconds.
    # Since we are only looking at specifically annotated Envoy sidecar containers
    # with a known metrics surface area and retain the values for <= 24 hours, we
    # can scrape aggressively. The higher scrape resolution is helpful for testing
    # and running checks that verify configuration health.
    global:
      scrape_interval: 5s
      scrape_timeout: 5s
      evaluation_interval: 5s

    # Scrape the Envoy sidecar metrics based on matching annotations (see above)
    scrape_configs:
    - job_name: 'opsani-envoy-sidecars'

      # Configure access to Kubernetes API server
      tls_config:
        ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
      bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token

      kubernetes_sd_configs:
        - role: pod
          namespaces:
            names:
            - {{ .Namespace }}

      relabel_configs:
        - action: labelmap
          regex: __meta_kubernetes_pod_label_(.+)
        - source_labels: [__meta_kubernetes_namespace]
          action: replace
          target_label: kubernetes_namespace
        - source_labels: [__meta_kubernetes_pod_name]
          action: replace
          target_label: kubernetes_pod_name

        # Do not attempt to scrape init containers
        - source_labels: [__meta_kubernetes_pod_container_init]
          action: drop
          regex: true

        # Only scrape Pods labeled with our optimizer
        - source_labels: [__meta_kubernetes_pod_label_servo_opsani_com_optimizer]
          action: keep
          regex: dev\.opsani\.com_leviathan-wakes

        # Relabel to scrape only pods that have
        # "prometheus.opsani.com/scrape = true" annotation.
        - source_labels: [__meta_kubernetes_pod_annotation_prometheus_opsani_com_scrape]
          action: keep
          regex: true

        # Relabel to configure scrape scheme for pod scrape targets
        # based on pod "prometheus.opsani.com/scheme = <scheme>" annotation.
        - source_labels: [__meta_kubernetes_pod_annotation_prometheus_opsani_com_scrape_scheme]
          action: replace
          target_label: __scheme__
          regex: (https?)

        # Relabel to customize metric path based on pod
        # "prometheus.opsani.com/path = <metric path>" annotation.
        - source_labels: [__meta_kubernetes_pod_annotation_prometheus_opsani_com_path]
          action: replace
          target_label: __metrics_path__
          regex: (.+)

        # Relabel to scrape only single, desired port for the pod
        # based on pod "prometheus.opsani.com/port = <port>" annotation.
        - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_opsani_com_port]
          action: replace
          regex: ([^:]+)(?::\d+)?;(\d+)
          replacement: $1:$2
          target_label: __address__

    - job_name: 'kubernetes-cadvisor'

      # Default to scraping over https. If required, just disable this or change to
      # http.
      scheme: https

      # Starting Kubernetes 1.7.3 the cAdvisor metrics are under /metrics/cadvisor.
      # Kubernetes CIS Benchmark recommends against enabling the insecure HTTP
      # servers of Kubernetes, therefore the cAdvisor metrics on the secure handler
      # are used.
      metrics_path: /metrics/cadvisor

      # This TLS & authorization config is used to connect to the actual scrape
      # endpoints for cluster components. This is separate to discovery auth
      # configuration because discovery & scraping are two separate concerns in
      # Prometheus. The discovery auth config is automatic if Prometheus runs inside
      # the cluster. Otherwise, more config options have to be provided within the
      # <kubernetes_sd_config>.
      tls_config:
        ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
        # If your node certificates are self-signed or use a different CA to the
        # master CA, then disable certificate verification below. Note that
        # certificate verification is an integral part of a secure infrastructure
        # so this should only be disabled in a controlled environment. You can
        # disable certificate verification by uncommenting the line below.
        #
        # insecure_skip_verify: true
      bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token

      kubernetes_sd_configs:
      - role: node

      relabel_configs:
      - action: labelmap
        regex: __meta_kubernetes_node_label_(.+)
`

// Configuration keys (Cobra and Viper)
const (
	KeyBaseURL        = "base-url"
	KeyOptimizer      = "optimizer"
	KeyToken          = "token"
	KeyProfile        = "profile"
	KeyDebugMode      = "debug"
	KeyRequestTracing = "trace-requests"
	KeyEnvPrefix      = "OPSANI"

	DefaultBaseURL = "https://api.opsani.com/"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
	BuiltBy   = "unknown"
)

func changelogURL(version string) string {
	path := "https://github.com/opsani/cli"
	r := regexp.MustCompile(`^v?\d+\.\d+\.\d+(-[\w.]+)?$`)
	if !r.MatchString(version) {
		return fmt.Sprintf("%s/releases/latest", path)
	}

	url := fmt.Sprintf("%s/releases/tag/v%s", path, strings.TrimPrefix(version, "v"))
	return url
}

// NewRootCommand returns a new instance of the root command for Opsani CLI
func NewRootCommand() *BaseCommand {
	// Create our base command to bind configuration
	viperCfg := viper.New()
	rootCmd := &BaseCommand{viperCfg: viperCfg}

	cobraCmd := &cobra.Command{
		Use:   "opsani",
		Short: "The official CLI for Opsani",
		Long: `Continuous optimization at your fingertips.

Opsani CLI is in early stages of development.
We'd love to hear your feedback at <https://github.com/opsani/cli>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		SilenceUsage:          true,
		SilenceErrors:         true,
		Version:               "0.1.2",
		DisableFlagsInUseLine: true,
	}

	// Link our root command to Cobra
	rootCmd.rootCobraCommand = cobraCmd

	// Set up versioning
	if Version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "(devel)" {
			Version = info.Main.Version
		}
	}
	Version = strings.TrimPrefix(Version, "v")
	if BuildDate == "" {
		cobraCmd.Version = Version
	} else {
		cobraCmd.Version = fmt.Sprintf("%s (%s)", Version, BuildDate)
	}
	versionOutput := fmt.Sprintf("Opsani CLI version %s\n%s\n", cobraCmd.Version, changelogURL(Version))
	cobraCmd.SetVersionTemplate(versionOutput)

	// Bind our global configuration parameters
	cobraCmd.PersistentFlags().String(KeyBaseURL, "", "Base URL for accessing the Opsani API")
	cobraCmd.PersistentFlags().MarkHidden(KeyBaseURL)
	cobraCmd.PersistentFlags().String(KeyOptimizer, "", "Optimizer to manage (overrides config file and OPSANI_OPTIMIZER)")
	cobraCmd.PersistentFlags().String(KeyToken, "", "Token for API authentication (overrides config file and OPSANI_TOKEN)")

	// Not stored in Viper
	cobraCmd.PersistentFlags().BoolVarP(&rootCmd.debugModeEnabled, KeyDebugMode, "D", false, "Enable debug mode")
	cobraCmd.PersistentFlags().BoolVar(&rootCmd.requestTracingEnabled, KeyRequestTracing, false, "Enable request tracing")

	// Respect NO_COLOR from env to be a good sport
	// https://no-color.org/
	_, disableColors := os.LookupEnv("NO_COLOR")
	cobraCmd.PersistentFlags().BoolVar(&rootCmd.disableColors, "no-colors", disableColors, "Disable colorized output")

	configFileUsage := fmt.Sprintf("Location of config file (default \"%s\")", rootCmd.DefaultConfigFile())
	cobraCmd.PersistentFlags().StringVar(&rootCmd.configFile, "config", "", configFileUsage)
	cobraCmd.MarkPersistentFlagFilename("config", "*.yaml", "*.yml")
	cobraCmd.PersistentFlags().StringP(KeyProfile, "p", os.Getenv("OPSANI_PROFILE"), "Profile to use (sets optimizer, token, and servo)")
	cobraCmd.Flags().Bool("version", false, "Display version and exit")
	cobraCmd.PersistentFlags().Bool("help", false, "Display help and exit")
	cobraCmd.PersistentFlags().MarkHidden("help")
	cobraCmd.PersistentFlags().MarkShorthandDeprecated("help", "please use --help")

	cobraCmd.SetHelpCommand(&cobra.Command{
		Hidden: true,
	})

	// Add all sub-commands
	cobraCmd.AddCommand(NewInitCommand(rootCmd))
	cobraCmd.AddCommand(NewOptimizerCommand(rootCmd))
	cobraCmd.AddCommand(NewServoCommand(rootCmd))
	cobraCmd.AddCommand(NewProfileCommand(rootCmd))

	cobraCmd.AddCommand(NewConsoleCommand(rootCmd))
	cobraCmd.AddCommand(NewConfigCommand(rootCmd))
	cobraCmd.AddCommand(NewCompletionCommand(rootCmd))

	cobraCmd.AddCommand(NewIgniteCommand(rootCmd))

	// Usage and help layout
	cobra.AddTemplateFunc("hasSubCommands", hasSubCommands)
	cobra.AddTemplateFunc("hasManagementSubCommands", hasManagementSubCommands)
	cobra.AddTemplateFunc("operationSubCommands", operationSubCommands)
	cobra.AddTemplateFunc("managementSubCommands", managementSubCommands)
	cobra.AddTemplateFunc("wrappedFlagUsages", wrappedFlagUsages)

	cobra.AddTemplateFunc("hasOtherSubCommands", hasOtherSubCommands)
	cobra.AddTemplateFunc("otherSubCommands", otherSubCommands)

	cobra.AddTemplateFunc("hasEducationalSubCommands", hasEducationalSubCommands)
	cobra.AddTemplateFunc("educationalSubCommands", educationalSubCommands)

	cobra.AddTemplateFunc("hasRegistrySubCommands", hasRegistrySubCommands)
	cobra.AddTemplateFunc("registrySubCommands", registrySubCommands)

	cobraCmd.SetUsageTemplate(usageTemplate)
	cobraCmd.SetHelpTemplate(helpTemplate)
	// cobraCmd.SetFlagErrorFunc(FlagErrorFunc)
	cobraCmd.SetHelpCommand(helpCommand)

	// See Execute()
	cobraCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		if err == pflag.ErrHelp {
			return err
		}
		return &FlagError{Err: err}
	})

	// Load configuration before execution of every action
	cobraCmd.PersistentPreRunE = ReduceRunEFuncs(rootCmd.InitConfigRunE, rootCmd.RequireConfigFileFlagToExistRunE)

	return rootCmd
}

// FlagError is the kind of error raised in flag processing
type FlagError struct {
	Err error
}

func (fe FlagError) Error() string {
	return fe.Err.Error()
}

func (fe FlagError) Unwrap() error {
	return fe.Err
}

func subCommandPath(rootCmd *cobra.Command, cmd *cobra.Command) string {
	path := make([]string, 0, 1)
	currentCmd := cmd
	if rootCmd == cmd {
		return ""
	}
	for {
		path = append([]string{currentCmd.Name()}, path...)
		if currentCmd.Parent() == rootCmd {
			return strings.Join(path, " ")
		}
		currentCmd = currentCmd.Parent()
	}
}

// Execute is the entry point for executing all commands from main
// All commands with RunE will bubble errors back here
func Execute() (cmd *cobra.Command, err error) {
	connectoToKubernetes()

	rootCmd := NewRootCommand()
	cobraCmd := rootCmd.rootCobraCommand

	executedCmd, err := rootCmd.rootCobraCommand.ExecuteC()
	if err != nil {
		// Exit silently if the user bailed with control-c
		if errors.Is(err, terminal.InterruptErr) {
			return executedCmd, err
		}

		executedCmd.PrintErrf("%s: %s\n", executedCmd.Name(), err)

		// Display usage for invalid command and flag errors
		var flagError *FlagError
		if errors.As(err, &flagError) || strings.HasPrefix(err.Error(), "unknown command ") {
			if !strings.HasSuffix(err.Error(), "\n") {
				executedCmd.PrintErrln()
			}
			executedCmd.PrintErrln(executedCmd.UsageString())
		}
	}
	return cobraCmd, err
}

// RunFunc is a Cobra Run function
type RunFunc func(cmd *cobra.Command, args []string)

// RunEFunc is a Cobra Run function that returns an error
type RunEFunc func(cmd *cobra.Command, args []string) error

// ReduceRunFuncs reduces a list of Cobra run functions into a single aggregate run function
func ReduceRunFuncs(runFuncs ...RunFunc) RunFunc {
	return func(cmd *cobra.Command, args []string) {
		for _, runFunc := range runFuncs {
			runFunc(cmd, args)
		}
	}
}

// ReduceRunEFuncs reduces a list of Cobra run functions that return an error into a single aggregate run function
func ReduceRunEFuncs(runFuncs ...RunEFunc) RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		for _, runFunc := range runFuncs {
			if err := runFunc(cmd, args); err != nil {
				return err
			}
		}
		return nil
	}
}

// InitConfigRunE initializes client configuration and aborts execution if an error is encountered
func (baseCmd *BaseCommand) InitConfigRunE(cmd *cobra.Command, args []string) error {
	return baseCmd.initConfig()
}

// RequireConfigFileFlagToExistRunE aborts command execution with an error if the config file specified via a flag does not exist
func (baseCmd *BaseCommand) RequireConfigFileFlagToExistRunE(cmd *cobra.Command, args []string) error {
	if configFilePath, err := baseCmd.PersistentFlags().GetString("config"); err == nil {
		if configFilePath != "" {
			if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
				return fmt.Errorf("config file does not exist. Run %q and try again (%w)",
					"opsani init", err)
			}
		}
	} else {
		return err
	}
	return nil
}

// RequireInitRunE aborts command execution with an error if the client is not initialized
func (baseCmd *BaseCommand) RequireInitRunE(cmd *cobra.Command, args []string) error {
	if !baseCmd.IsInitialized() {
		return fmt.Errorf("command failed because client is not initialized. Run %q and try again", "opsani init")
	}

	return nil
}

func (baseCmd *BaseCommand) initConfig() error {
	if baseCmd.configFile != "" {
		baseCmd.viperCfg.SetConfigFile(baseCmd.configFile)
	} else {
		// Find Opsani config in home directory
		baseCmd.viperCfg.AddConfigPath(baseCmd.DefaultConfigPath())
		baseCmd.viperCfg.SetConfigName("config")
		baseCmd.viperCfg.SetConfigType(baseCmd.DefaultConfigType())
	}

	// Set up environment variables
	baseCmd.viperCfg.SetEnvPrefix(KeyEnvPrefix)
	baseCmd.viperCfg.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	baseCmd.viperCfg.AutomaticEnv()

	// Load the configuration
	if err := baseCmd.viperCfg.ReadInConfig(); err == nil {
		if _, err = baseCmd.LoadProfile(); err != nil {
			return err
		}
	} else {
		// Ignore config file not found or error
		var perr *os.PathError
		if !errors.As(err, &viper.ConfigFileNotFoundError{}) &&
			!errors.As(err, &perr) {
			return fmt.Errorf("error parsing configuration file: %w", err)
		}
	}

	core.DisableColor = baseCmd.disableColors

	return nil
}

func (vitalCommand *vitalCommand) newSpinner() *spinner.Spinner {
	s := spinner.New(spinner.CharSets[14], 150*time.Millisecond)
	s.Writer = vitalCommand.OutOrStdout()
	s.Color("bold", "blue")
	s.HideCursor = true
	return s
}

func (vitalCommand *vitalCommand) infoMessage(message string) string {
	c := color.New(color.FgHiBlue, color.Bold).SprintFunc()
	return fmt.Sprintf("%s  %s\n", c("ℹ"), message)
}

func (vitalCommand *vitalCommand) successMessage(message string) string {
	c := color.New(color.FgGreen, color.Bold).SprintFunc()
	return fmt.Sprintf("%s  %s\n", c("\u2713"), message)
}

func (vitalCommand *vitalCommand) failureMessage(message string) string {
	c := color.New(color.Bold, color.FgHiRed).SprintFunc()
	return fmt.Sprintf("%s  %s\n", c("\u2717"), message)
}

// Task describes a long-running task that may succeed or fail
type Task struct {
	Description string
	Success     string
	Failure     string
	Run         func() error
	RunW        func(w io.Writer) error
	RunV        func() (interface{}, error)
}

// RunTaskWithSpinnerStatus displays an animated spinner around the execution of the given func
func (vitalCommand *vitalCommand) RunTaskWithSpinner(task Task) (err error) {
	s := vitalCommand.newSpinner()
	s.Suffix = "  " + task.Description
	s.Start()
	var templateVars interface{}
	if task.RunV != nil {
		templateVars, err = task.RunV()
	} else if task.RunW != nil {
		err = task.RunW(s.Writer)
	} else {
		err = task.Run()
	}
	s.Stop()

	if err == nil {
		tmpl, err := template.New("").Parse(task.Success)
		successMessage := new(bytes.Buffer)
		err = tmpl.Execute(successMessage, templateVars)
		if err != nil {
			return err
		}
		fmt.Fprintf(s.Writer, vitalCommand.successMessage(string(successMessage.Bytes())))
	} else {
		fmt.Fprintf(s.Writer, vitalCommand.failureMessage(fmt.Sprintf("%s: %s", task.Failure, err)))
	}
	return err
}

// RunTask displays runs a task
func (vitalCommand *vitalCommand) RunTask(task Task) (err error) {
	w := vitalCommand.OutOrStdout()
	fmt.Fprintf(w, vitalCommand.infoMessage(task.Description))
	if task.RunW != nil {
		err = task.RunW(w)
	} else {
		err = task.Run()
	}
	if err == nil {
		fmt.Fprintf(w, vitalCommand.successMessage(task.Success))
	} else {
		fmt.Fprintf(w, vitalCommand.failureMessage(task.Failure))
	}
	return err
}

// NewAPIClient returns an Opsani API client configured using the active configuration
func (baseCmd *BaseCommand) NewAPIClient() *opsani.Client {
	c := opsani.NewClient().
		SetBaseURL(baseCmd.BaseURL()).
		SetApp(baseCmd.Optimizer()).
		SetAuthToken(baseCmd.AccessToken()).
		SetDebug(baseCmd.DebugModeEnabled())
	if baseCmd.RequestTracingEnabled() {
		c.EnableTrace()
	}

	// Set the output directory to pwd by default
	if dir, err := os.Getwd(); err == nil {
		c.SetOutputDirectory(dir)
	}
	return c
}

// GetBaseURLHostnameAndPort returns the hostname and port portion of Opsani base URL for summary display
func (baseCmd *BaseCommand) GetBaseURLHostnameAndPort() string {
	u, err := url.Parse(baseCmd.GetBaseURL())
	if err != nil {
		return baseCmd.GetBaseURL()
	}
	baseURLDescription := u.Hostname()
	if port := u.Port(); port != "" && port != "80" && port != "443" {
		baseURLDescription = baseURLDescription + ":" + port
	}
	return baseURLDescription
}

// DefaultConfigFile returns the full path to the default Opsani configuration file
func (baseCmd *BaseCommand) DefaultConfigFile() string {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return filepath.Join(home, ".opsani", "config.yaml")
}

// DefaultConfigPath returns the path to the directory storing the Opsani configuration file
func (baseCmd *BaseCommand) DefaultConfigPath() string {
	return filepath.Dir(baseCmd.DefaultConfigFile())
}

// DefaultConfigType returns the
func (baseCmd *BaseCommand) DefaultConfigType() string {
	return "yaml"
}

// GetBaseURL returns the Opsani API base URL
func (baseCmd *BaseCommand) GetBaseURL() string {
	return baseCmd.viperCfg.GetString(KeyBaseURL)
}

// GetAppComponents returns the organization name and app ID as separate path components
func (baseCmd *BaseCommand) GetOptimizerComponents() (orgSlug string, appSlug string) {
	app := baseCmd.Optimizer()
	org := filepath.Dir(app)
	appID := filepath.Base(app)
	return org, appID
}

// GetAllSettings returns all configuration settings
func (baseCmd *BaseCommand) GetAllSettings() map[string]interface{} {
	return baseCmd.viperCfg.AllSettings()
}

// IsInitialized returns a boolean value that indicates if the client has been initialized
func (baseCmd *BaseCommand) IsInitialized() bool {
	return baseCmd.Optimizer() != "" && baseCmd.AccessToken() != ""
}

var helpCommand = &cobra.Command{
	Use:               "help [command]",
	Short:             "Help about the command",
	PersistentPreRun:  func(cmd *cobra.Command, args []string) {},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {},
	RunE: func(c *cobra.Command, args []string) error {
		cmd, args, e := c.Root().Find(args)
		if cmd == nil || e != nil || len(args) > 0 {
			return fmt.Errorf("unknown help topic: %v", strings.Join(args, " "))
		}

		helpFunc := cmd.HelpFunc()
		helpFunc(cmd, args)
		return nil
	},
}

/// Help and usage

// FlagErrorFunc prints an error message which matches the format of the
// docker/cli/cli error messages
// func FlagErrorFunc(cmd *cobra.Command, err error) error {
// 	if err == nil {
// 		return nil
// 	}

// 	usage := ""
// 	if cmd.HasSubCommands() {
// 		usage = "\n\n" + cmd.UsageString()
// 	}
// 	return StatusError{
// 		Status:     fmt.Sprintf("%s\nSee '%s --help'.%s", err, cmd.CommandPath(), usage),
// 		StatusCode: 125,
// 	}
// }

func hasSubCommands(cmd *cobra.Command) bool {
	return len(operationSubCommands(cmd)) > 0
}

func hasManagementSubCommands(cmd *cobra.Command) bool {
	return len(managementSubCommands(cmd)) > 0
}

func operationSubCommands(cmd *cobra.Command) []*cobra.Command {
	cmds := []*cobra.Command{}
	for _, sub := range cmd.Commands() {
		// if isOtherCommand(sub) {
		if len(sub.Annotations) > 0 {
			continue
		}
		if sub.IsAvailableCommand() && !sub.HasSubCommands() {
			cmds = append(cmds, sub)
		}
	}
	return cmds
}

func wrappedFlagUsages(cmd *cobra.Command) string {
	width := 80
	if ws, err := term.GetWinsize(0); err == nil {
		width = int(ws.Width)
	}
	return cmd.Flags().FlagUsagesWrapped(width - 1)
}

func managementSubCommands(cmd *cobra.Command) []*cobra.Command {
	cmds := []*cobra.Command{}
	for _, sub := range cmd.Commands() {
		if isOtherCommand(sub) {
			continue
		}
		if sub.IsAvailableCommand() && sub.HasSubCommands() && len(sub.Annotations) == 0 {
			cmds = append(cmds, sub)
		}
	}
	return cmds
}

func hasOtherSubCommands(cmd *cobra.Command) bool {
	return len(otherSubCommands(cmd)) > 0
}

func otherSubCommands(cmd *cobra.Command) []*cobra.Command {
	cmds := []*cobra.Command{}
	for _, sub := range cmd.Commands() {
		if sub.IsAvailableCommand() && isOtherCommand(sub) {
			cmds = append(cmds, sub)
		}
	}
	return cmds
}

func hasEducationalSubCommands(cmd *cobra.Command) bool {
	return len(educationalSubCommands(cmd)) > 0
}

func educationalSubCommands(cmd *cobra.Command) []*cobra.Command {
	cmds := []*cobra.Command{}
	for _, sub := range cmd.Commands() {
		if sub.IsAvailableCommand() && isEducationalCommand(sub) {
			cmds = append(cmds, sub)
		}
	}
	return cmds
}

func isEducationalCommand(cmd *cobra.Command) bool {
	return cmd.Annotations["educational"] == "true"
}

func isOtherCommand(cmd *cobra.Command) bool {
	return cmd.Annotations["other"] == "true"
}

func hasRegistrySubCommands(cmd *cobra.Command) bool {
	return len(registrySubCommands(cmd)) > 0
}

func registrySubCommands(cmd *cobra.Command) []*cobra.Command {
	cmds := []*cobra.Command{}
	for _, sub := range cmd.Commands() {
		if sub.IsAvailableCommand() && isRegistryCommand(sub) {
			cmds = append(cmds, sub)
		}
	}
	return cmds
}

func isRegistryCommand(cmd *cobra.Command) bool {
	return cmd.Annotations["registry"] == "true"
}

var usageTemplate = `Usage:

{{- if not .HasSubCommands}}	{{.UseLine}}{{end}}
{{- if .HasSubCommands}}	{{ .CommandPath}}{{- if .HasAvailableFlags}} [OPTIONS]{{end}} COMMAND{{end}}

{{if ne .Long ""}}{{ .Long | trim }}{{ else }}{{ .Short | trim }}{{end}}

{{- if gt .Aliases 0}}

Aliases:
  {{.NameAndAliases}}

{{- end}}
{{- if .HasExample}}

Examples:
{{ .Example }}

{{- end}}
{{- if .HasAvailableLocalFlags}}

Options:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Options:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

{{- end}}
{{- if hasManagementSubCommands . }}

Core Commands:

{{- range managementSubCommands . }}
  {{rpad .Name .NamePadding }} {{.Short}}
{{- end}}
{{- end}}
{{- if hasRegistrySubCommands . }}

Registry Commands:

{{- range registrySubCommands . }}
  {{rpad .Name .NamePadding }} {{.Short}}
{{- end}}
{{- end}}
{{- if hasSubCommands .}}

Commands:

{{- range operationSubCommands . }}
  {{rpad .Name .NamePadding }} {{.Short}}
{{- end}}
{{- end}}

{{- if hasEducationalSubCommands . }}

Learning Commands:

{{- range educationalSubCommands . }}
  {{rpad .Name .NamePadding }} {{.Short}}
{{- end}}
{{- end}}
{{- if hasOtherSubCommands .}}

Other Commands:

{{- range otherSubCommands . }}
  {{rpad .Name .NamePadding }} {{.Short}}
{{- end}}
{{- end}}

{{- if .HasSubCommands }}

Run '{{.CommandPath}} COMMAND --help' for more information on a command.
{{- end}}
`

var helpTemplate = `
{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`

func connectoToKubernetes() {
	var kubeconfig *string
	if home := k8s_homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		fmt.Printf("Connecting to Kubernetes with kubeconfig: %s\n", filepath.Join(home, ".kube", "config"))
	} else {
		fmt.Printf("Loading from path\n")
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	api := clientset.CoreV1()
	namespaces, err := api.Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Printf("Discovered %d namespaces in the cluster:\n\n", len(namespaces.Items))

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t") // pad with tabs
	table.SetNoWhiteSpace(true)

	data := [][]string{}
	headers := []string{"NAME", "STATUS", "AGE"}
	for _, namespace := range namespaces.Items {
		row := []string{
			namespace.Name,
			string(namespace.Status.Phase),
			timeago.NoMax(timeago.English).Format(namespace.ObjectMeta.CreationTimestamp.Time),
		}
		data = append(data, row)
	}
	table.SetHeader(headers)

	table.AppendBulk(data)
	// table.Render()

	// fmt.Printf("\n\n")

	// Show intro text
	markdown :=
		`# Opsani Team

## Let's talk about your cloud costs

It's the worst kept secret in tech. We're all spending way more on infrastructure than is necessary.

But it's not our fault. Our applications have become too big and complicated to optimize.

Until now.

## Better living through machine learning...

Opsani utilizes state of the art machine learning technology to continuously optimize your applications for *cost* and *performance*.

## Getting started

To start optimizing, a servo must be deployed into your cluster.

A servo is a lightweight container that lets Opsani know what is going on in your application and applies recommended configurations
provided by the optimizer.

This app is designed to assist you in assembling and deploying a servo through the miracle of automation and sensible defaults.

The process looks like...

- [x] Register for Opsani
- [x] Read this doc
- [ ] Configure optimization
- [ ] Start optimizing

## Things to keep in mind

All software run and deployed is Open Source. Opsani supports manual and assisted integrations if you like to do things the hard way.

Over the next 15 minutes, we will gather details about your application, the deployment environment, and your optimization goals.

Once optimization is configured to your liking, a servo will be deployed alongside your application to manage the optimization.

As tasks are completed, artifacts will be generated and saved onto this workstation.

Everything is logged, you can be pause and resume at any time, and no changes are applied without your confirmation.

Once this is wrapped up, optimization will begin immediately and you can sit back and enjoy the show.`
	r, err := glamour.NewTermRenderer(
		// TODO: detect background color and pick either the default dark or light theme
		glamour.WithStandardStyle("dark"),
	)
	if err != nil {
		log.Fatal(err)
	}
	renderedMarkdown, err := r.Render(markdown)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Fprint(os.Stdout, renderedMarkdown)

	// Page this shit
	// Put terminal in interactive mode
	fd := int(os.Stdin.Fd())
	oldState, err := ssh_terminal.MakeRaw(fd)
	if err != nil {
		log.Fatal(err)
	}

	var pager io.WriteCloser
	cmd, pager, err := runPager()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprint(pager, renderedMarkdown)
	pager.Close()
	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}
	ssh_terminal.Restore(fd, oldState)

	confirmed := false
	prompt := &survey.Confirm{
		Message: "Ready to get started?",
	}
	survey.AskOne(prompt, &confirmed)
	if confirmed {
		fmt.Printf("\n💥 Let's do this thing.\n\n")
	} else {
		log.Fatal("Bailing.")
	}

	// Select Namespace
	namespaceNames := []string{}
	for _, namespace := range namespaces.Items {
		namespaceNames = append(namespaceNames, namespace.Name)
	}

	namespace := ""
	survey.AskOne(&survey.Select{
		Message: "What namespace does your application run in?",
		Options: namespaceNames,
	}, &namespace)

	// List Deployments in Namespace
	extensionsApi := clientset.AppsV1()
	deployments, err := extensionsApi.Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	deploymentNames := []string{}
	for _, deployment := range deployments.Items {
		deploymentNames = append(deploymentNames, deployment.Name)
	}

	deploymentName := ""
	survey.AskOne(&survey.Select{
		Message: "Which deployment is running your application?",
		Options: deploymentNames,
	}, &deploymentName)

	var deployment v1.Deployment
	for _, d := range deployments.Items {
		if d.Name == deploymentName {
			deployment = d
			break
		}
	}

	// List the containers in the Deployment
	containerNames := []string{}
	for _, container := range deployment.Spec.Template.Spec.Containers {
		containerNames = append(containerNames, container.Name)
	}

	containerName := ""
	survey.AskOne(&survey.Select{
		Message: "Which container do you want to optimize?",
		Options: containerNames,
	}, &containerName)

	// List services that match the Deployment
	services, err := api.Services(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	serviceNames := []string{}
	for _, service := range services.Items {
		serviceNames = append(serviceNames, service.Name)
	}

	serviceName := ""
	survey.AskOne(&survey.Select{
		Message: "What service is providing traffic to your application?",
		Options: serviceNames,
	}, &serviceName)

	// Render manifest
	templateVars := map[string]string{"Namespace": namespace, "Deployment": deploymentName, "Container": containerName, "Service": serviceName}

	tmpl, err := template.New("").Parse(manifestTemplate)
	renderedManifest := new(bytes.Buffer)
	err = tmpl.Execute(renderedManifest, templateVars)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Rendered manifest: %s", string(renderedManifest.Bytes()))

	sepYamlfiles := strings.Split(string(renderedManifest.Bytes()), "---")
	retVal := make([]runtime.Object, 0, len(sepYamlfiles))
	for _, f := range sepYamlfiles {
		if f == "\n" || f == "" {
			// ignore empty cases
			continue
		}

		decode := scheme.Codecs.UniversalDeserializer().Decode
		obj, groupVersionKind, err := decode([]byte(f), nil, nil)

		if err != nil {
			log.Println(fmt.Sprintf("Error while decoding YAML object. Err was: %s", err))
			continue
		}

		log.Printf("Parsed object type: %s", groupVersionKind.Kind)
		retVal = append(retVal, obj)
	}

	for _, obj := range retVal {
		log.Printf("Creating object: %v", obj)
		createObject(clientset, *config, obj)
	}

	// Boom we are ready to roll
	profileOption := ""
	bold := color.New(color.Bold).SprintFunc()
	boldBlue := color.New(color.FgHiBlue, color.Bold).SprintFunc()
	fmt.Fprintf(os.Stdout, "\n🔥 %s\n", boldBlue("We have ignition"))
	fmt.Fprintf(os.Stdout, "\n%s  Optimizing Container %s of Deployment %s in Namespace %s", color.HiBlueString("ℹ"), bold(deploymentName), bold(containerName), bold(namespace))
	fmt.Fprintf(os.Stdout, "\n%s  Ingress traffic is routed from Service %s\n", color.HiBlueString("ℹ"), bold(serviceName))
	fmt.Fprintf(os.Stdout, "\n%s  Servo running in Kubernetes %s\n", color.HiBlueString("ℹ"), bold("deployments/servo"))
	// fmt.Fprintf(os.Stdout, "%s  Servo attached to opsani profile %s\n", color.HiBlueString("ℹ"), bold(vitalCommand.profile.Name))
	fmt.Fprintf(os.Stdout, "%s  Manifests written to %s\n", color.HiBlueString("ℹ"), bold("./manifests"))
	fmt.Fprintf(os.Stdout,
		"\n%s  View ignite subcommands: `%s`\n"+
			"%s  View servo subcommands: `%s`\n"+
			"%s  Follow servo logs: `%s`\n"+
			"%s  Watch pod status: `%s`\n"+
			"%s  Open Opsani console: `%s`\n\n",
		color.HiGreenString("❯"), color.YellowString(fmt.Sprintf("opsani %signite --help", profileOption)),
		color.HiGreenString("❯"), color.YellowString(fmt.Sprintf("opsani %sservo --help", profileOption)),
		color.HiGreenString("❯"), color.YellowString(fmt.Sprintf("opsani %sservo logs -f", profileOption)),
		color.HiGreenString("❯"), color.YellowString("kubectl get pods --watch"),
		color.HiGreenString("❯"), color.YellowString(fmt.Sprintf("opsani %sconsole", profileOption)))
	fmt.Println(bold("Optimization results will begin reporting in the console shortly."))

	fmt.Println("")
	log.Fatal("done.")

	return

	for {
		pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("There are %d pods in the cluster: %v\n", len(pods.Items), pods)

		// Examples for error handling:
		// - Use helper functions like e.g. errors.IsNotFound()
		// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
		namespace := "default"
		pod := "example-xxxxx"
		_, err = clientset.CoreV1().Pods(namespace).Get(context.TODO(), pod, metav1.GetOptions{})
		if k8s_errors.IsNotFound(err) {
			fmt.Printf("Pod %s in namespace %s not found\n", pod, namespace)
		} else if statusError, isStatus := err.(*k8s_errors.StatusError); isStatus {
			fmt.Printf("Error getting pod %s in namespace %s: %v\n",
				pod, namespace, statusError.ErrStatus.Message)
		} else if err != nil {
			panic(err.Error())
		} else {
			fmt.Printf("Found pod %s in namespace %s\n", pod, namespace)
		}

		time.Sleep(10 * time.Second)
	}
}

func createObject(kubeClientset kubernetes.Interface, restConfig rest.Config, obj runtime.Object) error {
	// Create a REST mapper that tracks information about the available resources in the cluster.
	groupResources, err := restmapper.GetAPIGroupResources(kubeClientset.Discovery())
	if err != nil {
		return err
	}
	rm := restmapper.NewDiscoveryRESTMapper(groupResources)

	// Get some metadata needed to make the REST request.
	gvk := obj.GetObjectKind().GroupVersionKind()
	gk := schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}
	mapping, err := rm.RESTMapping(gk, gvk.Version)
	if err != nil {
		return err
	}

	_, err = meta.NewAccessor().Name(obj)
	if err != nil {
		return err
	}

	// Create a client specifically for creating the object.
	restClient, err := newRestClient(restConfig, mapping.GroupVersionKind.GroupVersion())
	if err != nil {
		return err
	}

	// Use the REST helper to create the object in the "default" namespace.
	restHelper := resource.NewHelper(restClient, mapping)
	newObj, err := restHelper.Create("default", false, obj) //, &metav1.CreateOptions{})
	if err != nil {
		return err
	}

	fmt.Printf("Got new object: %s", newObj)
	return nil
}

func newRestClient(restConfig rest.Config, gv schema.GroupVersion) (rest.Interface, error) {
	restConfig.ContentConfig = resource.UnstructuredPlusDefaultContentConfig()
	restConfig.GroupVersion = &gv
	if len(gv.Group) == 0 {
		restConfig.APIPath = "/api"
	} else {
		restConfig.APIPath = "/apis"
	}

	return rest.RESTClientFor(&restConfig)
}
