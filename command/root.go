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
	"github.com/docker/docker/pkg/term"
	"github.com/fatih/color"
	"github.com/mitchellh/go-homedir"
	"github.com/opsani/cli/opsani"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// "k8s.io/client-go/kubernetes"
	// "k8s.io/client-go/tools/clientcmd"
	"github.com/xeonx/timeago"
	k8s_homedir "k8s.io/client-go/util/homedir"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/viper"
)

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
	fmt.Printf("Discovered %d namespaces in the cluster:\n\n", len(namespaces.Items))

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
	table.Render()

	fmt.Printf("\n\n")

	// Select Namespace
	namespaceNames := []string{}
	for _, namespace := range namespaces.Items {
		namespaceNames = append(namespaceNames, namespace.Name)
	}

	namespace := ""
	prompt := &survey.Select{
		Message: "Select Namespace:",
		Options: namespaceNames,
	}
	survey.AskOne(prompt, &namespace)

	// List Deployments in Namespace
	fmt.Printf("Looking for Deployments in Namespace %s\n", namespace)
	extensionsApi := clientset.AppsV1()
	deployments, err := extensionsApi.Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Discovered %d Deployment in the %s Namespace:\n\n", len(deployments.Items), namespace)

	deploymentNames := []string{}
	for _, deployment := range deployments.Items {
		deploymentNames = append(deploymentNames, deployment.Name)
	}

	deploymentName := ""
	survey.AskOne(&survey.Select{
		Message: "Select Deployment:",
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
		Message: "Select Container:",
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
		Message: "Select Service:",
		Options: serviceNames,
	}, &serviceName)

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
