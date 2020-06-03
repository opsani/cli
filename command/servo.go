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
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/AlecAivazis/survey/v2"
	"github.com/creack/pty"
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
	addCmd := &cobra.Command{
		Use:                   "add [OPTIONS] [NAME]",
		Long:                  "Add a servo to the local registry",
		Annotations:           map[string]string{"registry": "true"},
		Short:                 "Add a servo",
		Args:                  cobra.MaximumNArgs(1),
		RunE:                  servoCommand.RunAddServo,
		DisableFlagsInUseLine: true,
	}
	addCmd.Flags().BoolP("bastion", "b", false, "Use a bastion host for access")
	addCmd.Flags().String("bastion-host", "", "Specify the bastion host (format is user@host[:port])")
	servoCmd.AddCommand(addCmd)

	removeCmd := &cobra.Command{
		Use:                   "remove [OPTIONS] [NAME]",
		Long:                  "Remove a servo from the local registry",
		Annotations:           map[string]string{"registry": "true"},
		Aliases:               []string{"rm"},
		Short:                 "Remove a servo",
		Args:                  cobra.ExactArgs(1),
		RunE:                  servoCommand.RunRemoveServo,
		DisableFlagsInUseLine: true,
	}
	removeCmd.Flags().BoolVarP(&servoCommand.force, "force", "f", false, "Don't prompt for confirmation")
	servoCmd.AddCommand(removeCmd)

	// Servo Lifecycle
	servoCmd.AddCommand(&cobra.Command{
		Use:   "status [NAME]",
		Short: "Check servo status",
		Args:  cobra.ExactArgs(1),
		RunE:  servoCommand.RunServoStatus,
	})
	servoCmd.AddCommand(&cobra.Command{
		Use:   "start [NAME]",
		Short: "Start the servo",
		Args:  cobra.ExactArgs(1),
		RunE:  servoCommand.RunServoStart,
	})
	servoCmd.AddCommand(&cobra.Command{
		Use:   "stop [NAME]",
		Short: "Stop the servo",
		Args:  cobra.ExactArgs(1),
		RunE:  servoCommand.RunServoStop,
	})
	servoCmd.AddCommand(&cobra.Command{
		Use:   "restart [NAME]",
		Short: "Restart the servo",
		Args:  cobra.ExactArgs(1),
		RunE:  servoCommand.RunServoRestart,
	})

	// Servo Access
	servoCmd.AddCommand(&cobra.Command{
		Use:   "config [NAME]",
		Short: "View servo config file",
		Args:  cobra.ExactArgs(1),
		RunE:  servoCommand.RunServoConfig,
	})
	logsCmd := &cobra.Command{
		Use:   "logs [NAME]",
		Short: "View servo logs",
		Args:  cobra.ExactArgs(1),
		RunE:  servoCommand.RunServoLogs,
	}

	logsCmd.Flags().BoolVarP(&servoCommand.follow, "follow", "f", false, "Follow log output")
	logsCmd.Flags().BoolVarP(&servoCommand.timestamps, "timestamps", "t", false, "Show timestamps")
	logsCmd.Flags().StringVarP(&servoCommand.lines, "lines", "l", "25", `Number of lines to show from the end of the logs (or "all").`)

	servoCmd.AddCommand(logsCmd)
	servoCmd.AddCommand(&cobra.Command{
		Use:   "shell [NAME]",
		Short: "Open an interactive shell on the servo",
		Args:  cobra.ExactArgs(1),
		RunE:  servoCommand.RunServoShell,
	})

	return servoCmd
}

func (servoCmd *servoCommand) RunAddServo(c *cobra.Command, args []string) error {
	servo := Servo{}
	if len(args) > 0 {
		servo.Name = args[0]
	}

	if servo.Name == "" {
		err := servoCmd.AskOne(&survey.Input{
			Message: "Servo name?",
		}, &servo.Name, survey.WithValidator(survey.Required))
		if err != nil {
			return err
		}
	}

	if servo.Type == "" {
		err := servoCmd.AskOne(&survey.Select{
			Message: "Choose a color:",
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
				Message: "Namespace?",
				Default: "opsani",
			}, &servo.Namespace, survey.WithValidator(survey.Required))
			if err != nil {
				return err
			}
		}

		if servo.Host == "" {
			err := servoCmd.AskOne(&survey.Input{
				Message: "Deployment?",
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

	registry := NewServoRegistry(servoCmd.viperCfg)
	return registry.AddServo(servo)
}

func (servoCmd *servoCommand) RunRemoveServo(_ *cobra.Command, args []string) error {
	registry := NewServoRegistry(servoCmd.viperCfg)
	name := args[0]
	servo := registry.ServoNamed(name)
	if servo == nil {
		return fmt.Errorf("Unable to find servo %q", name)
	}

	confirmed := servoCmd.force
	if !confirmed {
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Remove servo %q?", servo.Name),
		}
		servoCmd.AskOne(prompt, &confirmed)
	}

	if confirmed {
		return registry.RemoveServo(*servo)
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
	registry := NewServoRegistry(servoCmd.viperCfg)
	servos, _ := registry.Servos()

	if servoCmd.verbose {
		headers := []string{"NAME", "TYPE", "NAMESPACE", "DEPLOYMENT", "USER", "HOST", "PATH"}
		for _, servo := range servos {
			row := []string{
				servo.Name,
				servo.Type,
				servo.Namespace,
				servo.Deployment,
				servo.User,
				servo.DisplayHost(),
				servo.DisplayPath(),
			}
			if servo.Bastion != "" {
				row = append(row, servo.Bastion)
				if len(headers) == 4 {
					headers = append(headers, "BASTION")
				}
			}
			data = append(data, row)
		}
		table.SetHeader(headers)
	} else {
		for _, servo := range servos {
			row := []string{
				servo.Name,
				servo.Type,
				servo.Description(),
			}
			if servo.Bastion != "" {
				row = append(row, fmt.Sprintf("(via %s)", servo.Bastion))
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

// Shell establishes an interactive shell with the servo
func (c *KubernetesServoDriver) Shell() error {
	argsS := fmt.Sprintf("-n %v exec -it deployment/%v -- /bin/bash", c.servo.Namespace, c.servo.Deployment)
	cmd := exec.Command("kubectl", ArgsS(argsS)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start the command with a pty.
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return err
	}
	// Make sure to close the pty at the end.
	defer func() { _ = ptmx.Close() }() // Best effort.

	// Handle pty size.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
				log.Printf("error resizing pty: %s", err)
			}
		}
	}()
	ch <- syscall.SIGWINCH // Initial resize.

	// Set stdin in raw mode.
	oldState, err := terminal.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	defer func() { _ = terminal.Restore(int(os.Stdin.Fd()), oldState) }() // Best effort.

	// Copy stdin to the pty and the pty to stdout.
	go func() { _, _ = io.Copy(ptmx, os.Stdin) }()
	_, err = io.Copy(os.Stdout, ptmx)
	return err
}

// NewServoCommander creates and returns an appropriate commander for a given servo
func NewServoCommander(servo Servo) (ServoDriver, error) {
	if servo.Type == "docker-compose" {
		return &DockerComposeServoDriver{servo: servo}, nil
	} else if servo.Type == "kubernetes" {
		return &KubernetesServoDriver{servo: servo}, nil
	}
	return nil, fmt.Errorf("no driver for servo %q (type %q)", servo.Name, servo.Type)
}

func (servoCmd *servoCommand) driverForServoNamed(name string) (ServoDriver, error) {
	registry := NewServoRegistry(servoCmd.viperCfg)
	servo := registry.ServoNamed(name)
	if servo == nil {
		return nil, fmt.Errorf("no such servo %q", name)
	}

	return NewServoCommander(*servo)
}

func (servoCmd *servoCommand) RunServoStatus(_ *cobra.Command, args []string) error {
	commander, err := servoCmd.driverForServoNamed(args[0])
	if commander == nil {
		return err
	}
	return commander.Status()
}

func (servoCmd *servoCommand) RunServoStart(_ *cobra.Command, args []string) error {
	commander, err := servoCmd.driverForServoNamed(args[0])
	if commander == nil {
		return err
	}
	return commander.Start()
}

func (servoCmd *servoCommand) RunServoStop(_ *cobra.Command, args []string) error {
	commander, err := servoCmd.driverForServoNamed(args[0])
	if commander == nil {
		return err
	}
	return commander.Stop()
}

func (servoCmd *servoCommand) RunServoRestart(_ *cobra.Command, args []string) error {
	commander, err := servoCmd.driverForServoNamed(args[0])
	if commander == nil {
		return err
	}
	return commander.Restart()
}

func (servoCmd *servoCommand) RunServoConfig(_ *cobra.Command, args []string) error {
	commander, err := servoCmd.driverForServoNamed(args[0])
	if commander == nil {
		return err
	}
	return commander.Config()
}

func (servoCmd *servoCommand) RunServoLogs(_ *cobra.Command, args []string) error {
	commander, err := servoCmd.driverForServoNamed(args[0])
	if commander == nil {
		return err
	}
	logsArgs := servoLogsArgs{
		Follow:     servoCmd.follow,
		Timestamps: servoCmd.timestamps,
		Lines:      servoCmd.lines,
	}
	return commander.Logs(logsArgs)
}

func (servoCmd *servoCommand) RunServoShell(_ *cobra.Command, args []string) error {
	commander, err := servoCmd.driverForServoNamed(args[0])
	if commander == nil {
		return err
	}
	return commander.Shell()
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
