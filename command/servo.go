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
		Use:   "status",
		Short: "Check servo status",
		Args:  cobra.ExactArgs(1),
		RunE:  servoCommand.RunServoStatus,
	})
	servoCmd.AddCommand(&cobra.Command{
		Use:   "start",
		Short: "Start the servo",
		Args:  cobra.ExactArgs(1),
		RunE:  servoCommand.RunServoStart,
	})
	servoCmd.AddCommand(&cobra.Command{
		Use:   "stop",
		Short: "Stop the servo",
		Args:  cobra.ExactArgs(1),
		RunE:  servoCommand.RunServoStop,
	})
	servoCmd.AddCommand(&cobra.Command{
		Use:   "restart",
		Short: "Restart servo",
		Args:  cobra.ExactArgs(1),
		RunE:  servoCommand.RunServoRestart,
	})

	// Servo Access
	servoCmd.AddCommand(&cobra.Command{
		Use:   "config",
		Short: "Display the servo config file",
		Args:  cobra.ExactArgs(1),
		RunE:  servoCommand.RunServoConfig,
	})
	logsCmd := &cobra.Command{
		Use:   "logs",
		Short: "View logs on a servo",
		Args:  cobra.ExactArgs(1),
		RunE:  servoCommand.RunServoLogs,
	}

	logsCmd.Flags().BoolVarP(&servoCommand.follow, "follow", "f", false, "Follow log output")
	logsCmd.Flags().BoolVarP(&servoCommand.timestamps, "timestamps", "t", false, "Show timestamps")
	logsCmd.Flags().StringVarP(&servoCommand.lines, "lines", "l", "25", `Number of lines to show from the end of the logs (or "all").`)

	servoCmd.AddCommand(logsCmd)
	servoCmd.AddCommand(&cobra.Command{
		Use:   "ssh",
		Short: "SSH into a servo",
		Args:  cobra.ExactArgs(1),
		RunE:  servoCommand.RunServoSSH,
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
		headers := []string{"NAME", "USER", "HOST", "PATH"}
		for _, servo := range servos {
			row := []string{
				servo.Name,
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
				servo.URL(),
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

func (servoCmd *servoCommand) RunServoStatus(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	return servoCmd.runInSSHSession(ctx, args[0], func(ctx context.Context, servo Servo, session *ssh.Session) error {
		return servoCmd.runDockerComposeOverSSH("ps", nil, servo, session)
	})
}

func (servoCmd *servoCommand) RunServoStart(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	return servoCmd.runInSSHSession(ctx, args[0], func(ctx context.Context, servo Servo, session *ssh.Session) error {
		return servoCmd.runDockerComposeOverSSH("up -d", nil, servo, session)
	})
}

func (servoCmd *servoCommand) RunServoStop(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	return servoCmd.runInSSHSession(ctx, args[0], func(ctx context.Context, servo Servo, session *ssh.Session) error {
		return servoCmd.runDockerComposeOverSSH("down", nil, servo, session)
	})
}

func (servoCmd *servoCommand) RunServoRestart(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	return servoCmd.runInSSHSession(ctx, args[0], func(ctx context.Context, servo Servo, session *ssh.Session) error {
		return servoCmd.runDockerComposeOverSSH("down && docker-compse up -d", nil, servo, session)
	})
}

func (servoCmd *servoCommand) RunServoConfig(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	outputBuffer := new(bytes.Buffer)
	err := servoCmd.runInSSHSession(ctx, args[0], func(ctx context.Context, servo Servo, session *ssh.Session) error {
		session.Stdout = outputBuffer
		session.Stderr = os.Stderr

		sshCmd := make([]string, 3)
		if path := servo.Path; path != "" {
			sshCmd = append(sshCmd, "cd", path+"&&")
		}
		sshCmd = append(sshCmd, "cat", "config.yaml")
		return session.Run(strings.Join(sshCmd, " "))
	})

	// We got the config, let's pretty print it
	if err == nil {
		servoCmd.PrettyPrintYAML(outputBuffer.Bytes(), true)
	}
	return err
}

func (servoCmd *servoCommand) RunServoLogs(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	return servoCmd.runInSSHSession(ctx, args[0], servoCmd.runLogsSSHSession)
}

// RunConfig displays Opsani CLI config info
func (servoCmd *servoCommand) RunServoSSH(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	return servoCmd.runInSSHSession(ctx, args[0], servoCmd.runShellOnSSHSession)
}

func (servoCmd *servoCommand) runDockerComposeOverSSH(cmd string, args []string, servo Servo, session *ssh.Session) error {
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	if path := servo.Path; path != "" {
		args = append(args, "cd", path+"&&")
	}
	args = append(args, "docker-compose", cmd)
	return session.Run(strings.Join(args, " "))
}

func (servoCmd *servoCommand) runLogsSSHSession(ctx context.Context, servo Servo, session *ssh.Session) error {
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	args := []string{}
	if path := servo.Path; path != "" {
		args = append(args, "cd", path+"&&")
	}
	args = append(args, "docker-compose logs")
	args = append(args, "--tail "+servoCmd.lines)
	if servoCmd.follow {
		args = append(args, "--follow")
	}
	if servoCmd.timestamps {
		args = append(args, "--timestamps")
	}
	return session.Run(strings.Join(args, " "))
}

func (servoCmd *servoCommand) runShellOnSSHSession(ctx context.Context, servo Servo, session *ssh.Session) error {
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

///
/// SSH Primitives
///

func (servoCmd *servoCommand) sshAgent() ssh.AuthMethod {
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
	}
	return nil
}

func (servoCmd *servoCommand) runInSSHSession(ctx context.Context, name string, runIt func(context.Context, Servo, *ssh.Session) error) error {
	registry := NewServoRegistry(servoCmd.viperCfg)
	servo := registry.ServoNamed(name)
	if servo == nil {
		return fmt.Errorf("no such servo %q", name)
	}

	// SSH client config
	knownHosts, err := homedir.Expand("~/.ssh/known_hosts")
	if err != nil {
		return err
	}
	hostKeyCallback, err := knownhosts.New(knownHosts)
	if err != nil {
		return err
	}
	config := &ssh.ClientConfig{
		User: servo.User,
		Auth: []ssh.AuthMethod{
			servoCmd.sshAgent(),
		},
		HostKeyCallback: hostKeyCallback,
	}

	// Support bastion hosts via redialing
	var sshClient *ssh.Client
	if servo.Bastion != "" {
		user, host := servo.BastionComponents()
		bastionConfig := &ssh.ClientConfig{
			User: user,
			Auth: []ssh.AuthMethod{
				servoCmd.sshAgent(),
			},
			HostKeyCallback: hostKeyCallback,
		}

		// Dial the bastion host
		bastionClient, err := ssh.Dial("tcp", host, bastionConfig)
		if err != nil {
			return err
		}

		// Establish a new connection thrrough the bastion
		conn, err := bastionClient.Dial("tcp", servo.HostAndPort())
		if err != nil {
			return err
		}

		// Build a new SSH connection on top of the bastion connection
		ncc, chans, reqs, err := ssh.NewClientConn(conn, servo.HostAndPort(), config)
		if err != nil {
			return err
		}

		// Now connection a client on top of it
		sshClient = ssh.NewClient(ncc, chans, reqs)
	} else {
		sshClient, err = ssh.Dial("tcp", servo.HostAndPort(), config)
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

	return runIt(ctx, *servo, session)
}
