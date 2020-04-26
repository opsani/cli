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
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/mitchellh/go-homedir"
	"github.com/olekukonko/tablewriter"
	"github.com/prometheus/common/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
	"golang.org/x/crypto/ssh/terminal"
)

type servoCommand struct {
	*BaseCommand
	force   bool
	verbose bool
}

// NewServoCommand returns a new instance of the servo command
func NewServoCommand(baseCmd *BaseCommand) *cobra.Command {
	servoCommand := servoCommand{BaseCommand: baseCmd}

	servoCmd := &cobra.Command{
		Use:   "servo",
		Short: "Manage Servos",
		Args:  cobra.NoArgs,
		PersistentPreRunE: ReduceRunEFuncs(
			baseCmd.InitConfigRunE,
			baseCmd.RequireConfigFileFlagToExistRunE,
			baseCmd.RequireInitRunE,
		),
	}

	// Servo registry
	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List Servos",
		Args:    cobra.NoArgs,
		RunE:    servoCommand.RunServoList,
	}
	listCmd.Flags().BoolVarP(&servoCommand.verbose, "verbose", "v", false, "Display verbose output")
	servoCmd.AddCommand(listCmd)
	servoCmd.AddCommand(&cobra.Command{
		Use:   "add",
		Short: "Add a Servo",
		Args:  cobra.MaximumNArgs(1),
		RunE:  servoCommand.RunAddServo,
	})

	removeCmd := &cobra.Command{
		Use:     "remove",
		Aliases: []string{"rm"},
		Short:   "Remove a Servo",
		Args:    cobra.ExactArgs(1),
		RunE:    servoCommand.RunRemoveServo,
	}
	removeCmd.Flags().BoolVarP(&servoCommand.force, "force", "f", false, "Don't prompt for confirmation")
	servoCmd.AddCommand(removeCmd)

	// Servo Lifecycle
	servoCmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Check Servo status",
		Args:  cobra.ExactArgs(1),
		RunE:  servoCommand.RunServoStatus,
	})
	servoCmd.AddCommand(&cobra.Command{
		Use:   "start",
		Short: "Start the Servo",
		Args:  cobra.ExactArgs(1),
		RunE:  servoCommand.RunServoStart,
	})
	servoCmd.AddCommand(&cobra.Command{
		Use:   "stop",
		Short: "Stop the Servo",
		Args:  cobra.ExactArgs(1),
		RunE:  servoCommand.RunServoStop,
	})
	servoCmd.AddCommand(&cobra.Command{
		Use:   "restart",
		Short: "Restart Servo",
		Args:  cobra.ExactArgs(1),
		RunE:  servoCommand.RunServoRestart,
	})

	// Servo Access
	logsCmd := &cobra.Command{
		Use:   "logs",
		Short: "View logs on a Servo",
		Args:  cobra.ExactArgs(1),
		RunE:  servoCommand.RunServoLogs,
	}

	logsCmd.Flags().BoolP("follow", "f", false, "Follow log output")
	viper.BindPFlag("follow", logsCmd.Flags().Lookup("follow"))
	logsCmd.Flags().BoolP("timestamps", "t", false, "Show timestamps")
	viper.BindPFlag("timestamps", logsCmd.Flags().Lookup("timestamps"))
	logsCmd.Flags().StringP("lines", "l", "25", `Number of lines to show from the end of the logs (or "all").`)
	viper.BindPFlag("lines", logsCmd.Flags().Lookup("lines"))

	servoCmd.AddCommand(logsCmd)
	servoCmd.AddCommand(&cobra.Command{
		Use:   "ssh",
		Short: "SSH into a Servo",
		Args:  cobra.ExactArgs(1),
		RunE:  servoCommand.RunServoSSH,
	})

	return servoCmd
}

// const username = "root"
// const hostname = "3.93.217.12"
// const port = "22"

// const sshKey = `
// -----BEGIN OPENSSH PRIVATE KEY-----
// FAKE KEY
// -----END OPENSSH PRIVATE KEY-----
// `

// var servos = []map[string]string{
// 	{
// 		"name": "opsani-dev",
// 		"host": "3.93.217.12",
// 		"port": "22",
// 		"user": "root",
// 		"path": "/root/dev.opsani.com/blake/oco",
// 	},
// }

func (servoCmd *servoCommand) SSHAgent() ssh.AuthMethod {
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
	}
	return nil
}

// Servo represents a deployed Servo assembly running somewhere
type Servo struct {
	Name string
	User string
	Host string
	Port string
	Path string
}

func (s Servo) HostAndPort() string {
	h := s.Host
	p := s.Port
	if p == "" {
		p = "22"
	}
	return strings.Join([]string{h, p}, ":")
}

func (s Servo) DisplayHost() string {
	v := s.Host
	if s.Port != "" && s.Port != "22" {
		v = v + ":" + s.Port
	}
	return v
}

func (s Servo) DisplayPath() string {
	if s.Path != "" {
		return s.Path
	}
	return "~/"
}

func (s Servo) URL() string {
	pathComponent := ""
	if s.Path != "" {
		if s.Port != "" && s.Port != "22" {
			pathComponent = pathComponent + ":"
		}
		pathComponent = pathComponent + s.Path
	}
	return fmt.Sprintf("ssh://%s@%s:%s", s.User, s.DisplayHost(), pathComponent)
}

func (servoCmd *servoCommand) RunAddServo(_ *cobra.Command, args []string) error {
	servo := Servo{}
	if len(args) > 0 {
		servo.Name = args[0]
	}

	servos := make([]Servo, 0)
	err := viper.UnmarshalKey("servos", &servos)
	if err != nil {
		return err
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
		}, &servo.Path, survey.WithValidator(survey.Required))
		if err != nil {
			return err
		}
	}

	servos = append(servos, servo)
	viper.Set("servos", servos)
	if err := viper.WriteConfig(); err != nil {
		return err
	}

	return nil
}

// GetServos returns the Servos in the configuration
func (servoCmd *servoCommand) GetServos() ([]Servo, error) {
	servos := make([]Servo, 0)
	err := viper.UnmarshalKey("servos", &servos)
	if err != nil {
		return nil, err
	}
	return servos, nil
}

func (servoCmd *servoCommand) RunRemoveServo(_ *cobra.Command, args []string) error {
	name := args[0]
	var servo *Servo

	// Find the target
	servos, err := servoCmd.GetServos()
	if err != nil {
		return err
	}
	var index int
	for i, s := range servos {
		if s.Name == name {
			servo = &s
			index = i
			break
		}
	}

	if servo == nil {
		return fmt.Errorf("Unable to find Servo named %q", name)
	}

	confirmed := servoCmd.force
	if !confirmed {
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Remove Servo %q?", servo.Name),
		}
		servoCmd.AskOne(prompt, &confirmed)
	}

	if confirmed {
		servos = append(servos[:index], servos[index+1:]...)
		viper.Set("servos", servos)
		if err := viper.WriteConfig(); err != nil {
			return err
		}
	}

	return nil
}

func (servoCmd *servoCommand) runInSSHSession(ctx context.Context, name string, runIt func(context.Context, Servo, *ssh.Session) error) error {

	// TODO: Recover from passphrase error
	// // signer, err := ssh.ParsePrivateKey([]byte(sshKey))
	// signer, err := ssh.ParsePrivateKey([]byte(sshKey))
	// signer, err := ssh.ParsePrivateKeyWithPassphrase([]byte(sshKey), []byte("THIS_IS_NOT_A_PASSPHRASE"))
	// if err != nil {
	// 	log.Base().Fatal(err)
	// }
	// fmt.Printf("Got signer %+v\n\n", signer)
	// key, err := x509.ParsePKCS1PrivateKey(der)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// signer := ssh.NewSignerFromKey(key)

	var servo *Servo
	servos, err := servoCmd.GetServos()
	for _, s := range servos {
		if s.Name == name {
			servo = &s
			break
		}
	}
	if servo == nil {
		return fmt.Errorf("no such Servo %q", name)
	}

	// SSH client config
	knownHosts, err := homedir.Expand("~/.ssh/known_hosts")
	if err != nil {
		return err
	}
	hostKeyCallback, err := knownhosts.New(knownHosts)
	if err != nil {
		log.Fatal("could not create hostkeycallback function: ", err)
	}
	config := &ssh.ClientConfig{
		User: servo.User,
		// Auth: []ssh.AuthMethod{
		// 	// ssh.Password(password),
		// 	ssh.PublicKeys(signer),
		// },
		Auth: []ssh.AuthMethod{
			servoCmd.SSHAgent(),
		},
		// Non-production only
		HostKeyCallback: hostKeyCallback,
	}

	// // Connect to host
	fmt.Printf("Servos: %+v\nServo=%+v\n\n", servos, servo)
	client, err := ssh.Dial("tcp", servo.HostAndPort(), config)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Create sesssion
	session, err := client.NewSession()
	if err != nil {
		log.Fatal("Failed to create session: ", err)
	}
	defer session.Close()

	go func() {
		<-ctx.Done()
		client.Close()
	}()

	return runIt(ctx, *servo, session)
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
	servos, _ := servoCmd.GetServos()

	if servoCmd.verbose {
		table.SetHeader([]string{"NAME", "USER", "HOST", "PATH"})
		for _, servo := range servos {
			data = append(data, []string{
				servo.Name,
				servo.User,
				servo.DisplayHost(),
				servo.DisplayPath(),
			})
		}
	} else {
		for _, servo := range servos {
			data = append(data, []string{
				servo.Name,
				servo.URL(),
			})
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
		return servoCmd.runDockerComposeOverSSH("start", nil, servo, session)
	})
}

func (servoCmd *servoCommand) RunServoStop(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	return servoCmd.runInSSHSession(ctx, args[0], func(ctx context.Context, servo Servo, session *ssh.Session) error {
		return servoCmd.runDockerComposeOverSSH("stop", nil, servo, session)
	})
}

func (servoCmd *servoCommand) RunServoRestart(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	return servoCmd.runInSSHSession(ctx, args[0], func(ctx context.Context, servo Servo, session *ssh.Session) error {
		return servoCmd.runDockerComposeOverSSH("stop && docker-compse start", nil, servo, session)
	})
}

func (servoCmd *servoCommand) RunServoLogs(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	return servoCmd.runInSSHSession(ctx, args[0], servoCmd.RunLogsSSHSession)
}

// RunConfig displays Opsani CLI config info
func (servoCmd *servoCommand) RunServoSSH(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	return servoCmd.runInSSHSession(ctx, args[0], servoCmd.RunShellOnSSHSession)
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

func (servoCmd *servoCommand) RunLogsSSHSession(ctx context.Context, servo Servo, session *ssh.Session) error {
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	args := []string{}
	if path := servo.Path; path != "" {
		args = append(args, "cd", path+"&&")
	}
	args = append(args, "docker-compose logs")
	args = append(args, "--tail "+viper.GetString("lines"))
	if viper.GetBool("follow") {
		args = append(args, "--follow")
	}
	if viper.GetBool("timestamps") {
		args = append(args, "--timestamps")
	}
	return session.Run(strings.Join(args, " "))
}

func (servoCmd *servoCommand) RunShellOnSSHSession(ctx context.Context, servo Servo, session *ssh.Session) error {
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
