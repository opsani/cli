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
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/mitchellh/go-homedir"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
	"golang.org/x/crypto/ssh/terminal"
)

// Args is a convenience function that converts a variadic list of strings into an array
func Args(args ...string) []string {
	return args
}

// ArgsS is a convenience function that converts a space delimited string into an array of args
func ArgsS(args string) []string {
	return strings.Split(args, " ")
}

// NOTE: Binding vars instead of using flags because the call stack is messy atm
type servoCommand struct {
	*BaseCommand
	force      bool
	verbose    bool
	follow     bool
	timestamps bool
	lines      string
}

// NewServoCommand returns a new instance of the servo command
func NewServoCommand(baseCmd *BaseCommand) *cobra.Command {
	servoCommand := servoCommand{BaseCommand: baseCmd}

	servoCmd := &cobra.Command{
		Use:   "servo",
		Short: "Manage servos",
		Args:  cobra.NoArgs,
		PersistentPreRunE: ReduceRunEFuncs(
			baseCmd.InitConfigRunE,
			baseCmd.RequireConfigFileFlagToExistRunE,
			baseCmd.RequireInitRunE,
		),
	}

	// Servo registry
	listCmd := &cobra.Command{
		Use:         "list",
		Annotations: map[string]string{"registry": "true"},
		Aliases:     []string{"ls"},
		Short:       "List servos",
		Args:        cobra.NoArgs,
		RunE:        servoCommand.RunServoList,
	}
	listCmd.Flags().BoolVarP(&servoCommand.verbose, "verbose", "v", false, "Display verbose output")
	servoCmd.AddCommand(listCmd)
	attachCmd := &cobra.Command{
		Use:                   "attach [OPTIONS]",
		Long:                  "Attach servo to the active profile",
		Annotations:           map[string]string{"registry": "true"},
		Short:                 "Attach servo to active profile",
		Args:                  cobra.NoArgs,
		RunE:                  servoCommand.RunAttachServo,
		DisableFlagsInUseLine: true,
	}
	attachCmd.Flags().BoolP("bastion", "b", false, "Use a bastion host for access")
	attachCmd.Flags().String("bastion-host", "", "Specify the bastion host (format is user@host[:port])")
	servoCmd.AddCommand(attachCmd)

	detachCmd := &cobra.Command{
		Use:                   "detach [OPTIONS]",
		Long:                  "Detach servo from the active profile",
		Annotations:           map[string]string{"registry": "true"},
		Aliases:               []string{"rm"},
		Short:                 "Detach servo from active profile",
		Args:                  cobra.NoArgs,
		RunE:                  servoCommand.RunDetachServo,
		DisableFlagsInUseLine: true,
	}
	detachCmd.Flags().BoolVarP(&servoCommand.force, "force", "f", false, "Don't prompt for confirmation")
	servoCmd.AddCommand(detachCmd)

	// Servo Lifecycle
	servoCmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Check servo status",
		Args:  cobra.NoArgs,
		RunE:  servoCommand.RunServoStatus,
	})
	servoCmd.AddCommand(&cobra.Command{
		Use:   "start",
		Short: "Start the servo",
		Args:  cobra.NoArgs,
		RunE:  servoCommand.RunServoStart,
	})
	servoCmd.AddCommand(&cobra.Command{
		Use:   "stop",
		Short: "Stop the servo",
		Args:  cobra.NoArgs,
		RunE:  servoCommand.RunServoStop,
	})
	servoCmd.AddCommand(&cobra.Command{
		Use:   "restart",
		Short: "Restart the servo",
		Args:  cobra.NoArgs,
		RunE:  servoCommand.RunServoRestart,
	})

	// Servo Access
	servoCmd.AddCommand(&cobra.Command{
		Use:   "config",
		Short: "View servo config file",
		Args:  cobra.NoArgs,
		RunE:  servoCommand.RunServoConfig,
	})
	logsCmd := &cobra.Command{
		Use:   "logs",
		Short: "View servo logs",
		Args:  cobra.NoArgs,
		RunE:  servoCommand.RunServoLogs,
	}

	logsCmd.Flags().BoolVarP(&servoCommand.follow, "follow", "f", false, "Follow log output")
	logsCmd.Flags().BoolVarP(&servoCommand.timestamps, "timestamps", "t", false, "Show timestamps")
	logsCmd.Flags().StringVarP(&servoCommand.lines, "lines", "l", "25", `Number of lines to show from the end of the logs (or "all").`)

	servoCmd.AddCommand(logsCmd)
	servoCmd.AddCommand(&cobra.Command{
		Use:   "shell",
		Short: "Open an interactive shell on the servo",
		Args:  cobra.NoArgs,
		RunE:  servoCommand.RunServoShell,
	})

	return servoCmd
}

func (servoCmd *servoCommand) RunAttachServo(c *cobra.Command, args []string) error {
	if servoCmd.profile == nil {
		return fmt.Errorf("no profile active")
	}

	if servoCmd.profile.Servo != (Servo{}) {
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Existing servo attached to %q. Overwrite?", servoCmd.profile.Name),
		}
		var confirmed bool
		servoCmd.AskOne(prompt, &confirmed)
		if confirmed == false {
			return nil
		}
	}

	servo := Servo{}

	if servo.Type == "" {
		err := servoCmd.AskOne(&survey.Select{
			Message: "Select deployment:",
			Options: []string{"kubernetes", "docker-compose"},
			Default: "kubernetes",
		}, &servo.Type, survey.WithValidator(survey.Required))
		if err != nil {
			return err
		}
	}

	if servo.Type == "kubernetes" {
		if servo.User == "" {
			err := servoCmd.AskOne(&survey.Input{
				Message: "Namespace:",
				Default: "opsani",
			}, &servo.Namespace, survey.WithValidator(survey.Required))
			if err != nil {
				return err
			}
		}

		if servo.Host == "" {
			err := servoCmd.AskOne(&survey.Input{
				Message: "Deployment:",
				Default: "servo",
			}, &servo.Deployment, survey.WithValidator(survey.Required))
			if err != nil {
				return err
			}
		}
	}

	if servo.Type == "docker-compose" {
		if servo.User == "" {
			err := servoCmd.AskOne(&survey.Input{
				Message: "User?",
			}, &servo.User, survey.WithValidator(survey.Required))
			if err != nil {
				return err
			}
		}

		if servo.Host == "" {
			err := servoCmd.AskOne(&survey.Input{
				Message: "Host?",
			}, &servo.Host, survey.WithValidator(survey.Required))
			if err != nil {
				return err
			}
		}

		if servo.Path == "" {
			err := servoCmd.AskOne(&survey.Input{
				Message: "Path? (optional)",
			}, &servo.Path)
			if err != nil {
				return err
			}
		}

		// Handle bastion hosts
		if flagSet, _ := c.Flags().GetBool("bastion"); flagSet {
			servo.Bastion, _ = c.Flags().GetString("bastion-host")
			if servo.Bastion == "" {
				err := servoCmd.AskOne(&survey.Input{
					Message: "Bastion host? (format is user@host[:port])",
				}, &servo.Bastion)
				if err != nil {
					return err
				}
			}
		}
	}

	registry, err := NewProfileRegistry(servoCmd.viperCfg)
	if err != nil {
		return err
	}
	profile := registry.ProfileNamed(servoCmd.profile.Name)
	profile.Servo = servo
	if err := registry.Save(); err != nil {
		return err
	}

	return nil
}

func (servoCmd *servoCommand) RunDetachServo(_ *cobra.Command, args []string) error {
	if servoCmd.profile == nil {
		return fmt.Errorf("no profile active")
	} else if servoCmd.profile.Servo == (Servo{}) {
		return fmt.Errorf("no servo is attached")
	}

	confirmed := servoCmd.force
	if !confirmed {
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Detach servo from profile %q?", servoCmd.profile.Name),
		}
		servoCmd.AskOne(prompt, &confirmed)
	}

	if confirmed {
		registry, err := NewProfileRegistry(servoCmd.viperCfg)
		if err != nil {
			return err
		}
		profile := registry.ProfileNamed(servoCmd.profile.Name)
		profile.Servo = Servo{}
		if err := registry.Save(); err != nil {
			return err
		}
		// return registry.RemoveServo(*servo)
	}

	return nil
}

func (servoCmd *servoCommand) RunServoList(_ *cobra.Command, args []string) error {
	table := tablewriter.NewWriter(servoCmd.OutOrStdout())
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
	registry, err := NewProfileRegistry(servoCmd.viperCfg)
	if err != nil {
		return nil
	}

	if servoCmd.verbose {
		headers := []string{"NAME", "TYPE", "NAMESPACE", "DEPLOYMENT", "USER", "HOST", "PATH"}
		for _, profile := range registry.Profiles() {
			row := []string{
				profile.Name,
				profile.Servo.Type,
				profile.Servo.Namespace,
				profile.Servo.Deployment,
				profile.Servo.User,
				profile.Servo.DisplayHost(),
				profile.Servo.DisplayPath(),
			}
			if profile.Servo.Bastion != "" {
				row = append(row, profile.Servo.Bastion)
				if len(headers) == 4 {
					headers = append(headers, "BASTION")
				}
			}
			data = append(data, row)
		}
		table.SetHeader(headers)
	} else {
		for _, profile := range registry.Profiles() {
			row := []string{
				profile.Name,
				profile.Servo.Type,
				profile.Servo.Description(),
			}
			if profile.Servo.Bastion != "" {
				row = append(row, fmt.Sprintf("(via %s)", profile.Servo.Bastion))
			}
			data = append(data, row)
		}
	}

	table.AppendBulk(data)
	table.Render()
	return nil
}

type servoLogsArgs struct {
	Follow     bool
	Timestamps bool
	Lines      string
}

// ServoDriver defines a standard interface for interacting with servo deployments
type ServoDriver interface {
	Status() error // TODO: pass io.Writer for output, ssh interface for bastion
	Start() error
	Stop() error
	Restart() error
	Logs(args servoLogsArgs) error
	Config() error
	Shell() error
}

// DockerComposeServoDriver supports interaction with servos deployed via Docker Compose
type DockerComposeServoDriver struct {
	servo Servo
}

// Status outputs the servo status
func (c *DockerComposeServoDriver) Status() error {
	ctx := context.Background()
	return c.runInSSHSession(ctx, func(ctx context.Context, session *ssh.Session) error {
		return c.runDockerComposeOverSSH("ps", nil, session)
	})
}

// Start starts the servo
func (c *DockerComposeServoDriver) Start() error {
	ctx := context.Background()
	return c.runInSSHSession(ctx, func(ctx context.Context, session *ssh.Session) error {
		return c.runDockerComposeOverSSH("up -d", nil, session)
	})
}

// Stop stops the servo
func (c *DockerComposeServoDriver) Stop() error {
	ctx := context.Background()
	return c.runInSSHSession(ctx, func(ctx context.Context, session *ssh.Session) error {
		return c.runDockerComposeOverSSH("down", nil, session)
	})
}

// Restart restrarts the servo
func (c *DockerComposeServoDriver) Restart() error {
	ctx := context.Background()
	return c.runInSSHSession(ctx, func(ctx context.Context, session *ssh.Session) error {
		return c.runDockerComposeOverSSH("down && docker-compse up -d", nil, session)
	})
}

// Logs outputs the servo logs
func (c *DockerComposeServoDriver) Logs(logsArgs servoLogsArgs) error {
	ctx := context.Background()
	return c.runInSSHSession(ctx, func(ctx context.Context, session *ssh.Session) error {
		// TODO: Needs to be passed in
		session.Stdout = os.Stdout
		session.Stderr = os.Stderr

		args := []string{}
		if path := c.servo.Path; path != "" {
			args = append(args, "cd", path+"&&")
		}
		args = append(args, "docker-compose logs")
		args = append(args, "--tail "+logsArgs.Lines)
		if logsArgs.Follow {
			args = append(args, "--follow")
		}
		if logsArgs.Timestamps {
			args = append(args, "--timestamps")
		}
		return session.Run(strings.Join(args, " "))
	})
}

// Config returns the servo config file
func (c *DockerComposeServoDriver) Config() error {
	ctx := context.Background()
	outputBuffer := new(bytes.Buffer)
	err := c.runInSSHSession(ctx, func(ctx context.Context, session *ssh.Session) error {
		session.Stdout = outputBuffer
		session.Stderr = os.Stderr

		sshCmd := make([]string, 3)
		if path := c.servo.Path; path != "" {
			sshCmd = append(sshCmd, "cd", path+"&&")
		}
		sshCmd = append(sshCmd, "cat", "config.yaml")
		return session.Run(strings.Join(sshCmd, " "))
	})

	// We got the config, let's pretty print it
	if err == nil {
		prettyYAML, _ := PrettyPrintYAMLToString(outputBuffer.Bytes(), true, true)
		_, err = os.Stdout.Write([]byte(prettyYAML + "\n"))
	}
	return err
}

// Shell establishes an interactive shell with the servo
func (c *DockerComposeServoDriver) Shell() error {
	ctx := context.Background()
	return c.runInSSHSession(ctx, c.runShellOnSSHSession)
}

func (c *DockerComposeServoDriver) runDockerComposeOverSSH(cmd string, args []string, session *ssh.Session) error {
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	if path := c.servo.Path; path != "" {
		args = append(args, "cd", path+"&&")
	}
	args = append(args, "docker-compose", cmd)
	return session.Run(strings.Join(args, " "))
}

func (c *DockerComposeServoDriver) runShellOnSSHSession(ctx context.Context, session *ssh.Session) error {
	fd := int(os.Stdin.Fd())
	state, err := terminal.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("terminal make raw: %s", err)
	}
	defer terminal.Restore(fd, state)

	w, h, err := terminal.GetSize(fd)
	if err != nil {
		return fmt.Errorf("terminal get size: %s", err)
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	term := os.Getenv("TERM")
	if term == "" {
		term = "xterm-256color"
	}
	if err := session.RequestPty(term, h, w, modes); err != nil {
		return fmt.Errorf("session xterm: %s", err)
	}

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	if err := session.Shell(); err != nil {
		return fmt.Errorf("session shell: %s", err)
	}

	if err := session.Wait(); err != nil {
		if e, ok := err.(*ssh.ExitError); ok {
			switch e.ExitStatus() {
			case 130:
				return nil
			}
		}
		return fmt.Errorf("ssh: %s", err)
	}

	return err
}

//////////////////

// KubernetesServoDriver supports interaction with servos deployed via Kubernetes
type KubernetesServoDriver struct {
	servo Servo
}

// Status outputs the servo status
func (c *KubernetesServoDriver) Status() error {
	argsS := fmt.Sprintf("-n %v describe deployments/%v", c.servo.Namespace, c.servo.Deployment)
	cmd := exec.Command("kubectl", ArgsS(argsS)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Start starts the servo
func (c *KubernetesServoDriver) Start() error {
	argsS := fmt.Sprintf("-n %v scale --replicas=1 deployments/%v", c.servo.Namespace, c.servo.Deployment)
	cmd := exec.Command("kubectl", ArgsS(argsS)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Stop stops the servo
func (c *KubernetesServoDriver) Stop() error {
	argsS := fmt.Sprintf("-n %v scale --replicas=0 deployments/%v", c.servo.Namespace, c.servo.Deployment)
	cmd := exec.Command("kubectl", ArgsS(argsS)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Restart restarts the servo
func (c *KubernetesServoDriver) Restart() error {
	argsS := fmt.Sprintf("-n %v rollout restart deployment/%v", c.servo.Namespace, c.servo.Deployment)
	cmd := exec.Command("kubectl", ArgsS(argsS)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Logs outputs the servo logs
func (c *KubernetesServoDriver) Logs(logsArgs servoLogsArgs) error {
	deploymentArg := fmt.Sprintf("deployments/%v", c.servo.Deployment)
	args := Args("-n", c.servo.Namespace, "logs", deploymentArg)

	if logsArgs.Lines != "" {
		args = append(args, "--tail="+logsArgs.Lines)
	}
	if logsArgs.Follow {
		args = append(args, "--follow")
	}
	if logsArgs.Timestamps {
		args = append(args, "--timestamps")
	}

	cmd := exec.Command("kubectl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Config outputs the servo config
func (c *KubernetesServoDriver) Config() error {
	outputBuffer := new(bytes.Buffer)
	argsS := fmt.Sprintf("-n %v exec deployment/%v -- cat /servo/config.yaml", c.servo.Namespace, c.servo.Deployment)
	cmd := exec.Command("kubectl", ArgsS(argsS)...)
	cmd.Stdout = outputBuffer
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil
	}

	prettyYAML, _ := PrettyPrintYAMLToString(outputBuffer.Bytes(), true, true)
	_, err := os.Stdout.Write([]byte(prettyYAML + "\n"))
	return err
}

// NewServoDriver creates and returns an appropriate commander for a given servo
func NewServoDriver(servo Servo) (ServoDriver, error) {
	if servo.Type == "docker-compose" {
		return &DockerComposeServoDriver{servo: servo}, nil
	} else if servo.Type == "kubernetes" {
		return &KubernetesServoDriver{servo: servo}, nil
	}
	return nil, fmt.Errorf("no driver for servo type: %q", servo.Type)
}

func (servoCmd *servoCommand) RunServoStatus(_ *cobra.Command, args []string) error {
	driver, err := NewServoDriver(servoCmd.profile.Servo)
	if driver == nil {
		return err
	}
	return driver.Status()
}

func (servoCmd *servoCommand) RunServoStart(_ *cobra.Command, args []string) error {
	driver, err := NewServoDriver(servoCmd.profile.Servo)
	if driver == nil {
		return err
	}
	return driver.Start()
}

func (servoCmd *servoCommand) RunServoStop(_ *cobra.Command, args []string) error {
	driver, err := NewServoDriver(servoCmd.profile.Servo)
	if driver == nil {
		return err
	}
	return driver.Stop()
}

func (servoCmd *servoCommand) RunServoRestart(_ *cobra.Command, args []string) error {
	driver, err := NewServoDriver(servoCmd.profile.Servo)
	if driver == nil {
		return err
	}
	return driver.Restart()
}

func (servoCmd *servoCommand) RunServoConfig(_ *cobra.Command, args []string) error {
	driver, err := NewServoDriver(servoCmd.profile.Servo)
	if driver == nil {
		return err
	}
	return driver.Config()
}

func (servoCmd *servoCommand) RunServoLogs(_ *cobra.Command, args []string) error {
	driver, err := NewServoDriver(servoCmd.profile.Servo)
	if driver == nil {
		return err
	}
	logsArgs := servoLogsArgs{
		Follow:     servoCmd.follow,
		Timestamps: servoCmd.timestamps,
		Lines:      servoCmd.lines,
	}
	return driver.Logs(logsArgs)
}

func (servoCmd *servoCommand) RunServoShell(_ *cobra.Command, args []string) error {
	driver, err := NewServoDriver(servoCmd.profile.Servo)
	if driver == nil {
		return err
	}
	return driver.Shell()
}

///
/// SSH Primitives
///

func sshAgent() ssh.AuthMethod {
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
	}
	return nil
}

// TODO: convert to standalone func
func (c *DockerComposeServoDriver) runInSSHSession(ctx context.Context, runIt func(context.Context, *ssh.Session) error) error {
	// SSH client config
	knownHosts, err := homedir.Expand("~/.ssh/known_hosts") // TODO: Windows support
	if err != nil {
		return err
	}
	hostKeyCallback, err := knownhosts.New(knownHosts)
	if err != nil {
		return err
	}
	config := &ssh.ClientConfig{
		User: c.servo.User,
		Auth: []ssh.AuthMethod{
			sshAgent(),
		},
		HostKeyCallback: hostKeyCallback,
	}

	// Support bastion hosts via redialing
	var sshClient *ssh.Client
	if c.servo.Bastion != "" {
		user, host := c.servo.BastionComponents()
		bastionConfig := &ssh.ClientConfig{
			User: user,
			Auth: []ssh.AuthMethod{
				sshAgent(),
			},
			HostKeyCallback: hostKeyCallback,
		}

		// Dial the bastion host
		bastionClient, err := ssh.Dial("tcp", host, bastionConfig)
		if err != nil {
			return err
		}

		// Establish a new connection thrrough the bastion
		conn, err := bastionClient.Dial("tcp", c.servo.HostAndPort())
		if err != nil {
			return err
		}

		// Build a new SSH connection on top of the bastion connection
		ncc, chans, reqs, err := ssh.NewClientConn(conn, c.servo.HostAndPort(), config)
		if err != nil {
			return err
		}

		// Now connection a client on top of it
		sshClient = ssh.NewClient(ncc, chans, reqs)
	} else {
		sshClient, err = ssh.Dial("tcp", c.servo.HostAndPort(), config)
		if err != nil {
			return err
		}
	}
	defer sshClient.Close()

	// Create sesssion
	session, err := sshClient.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	go func() {
		<-ctx.Done()
		sshClient.Close()
	}()

	return runIt(ctx, session)
}
