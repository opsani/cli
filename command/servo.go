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
	servoCmd.AddCommand(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List Servos",
		Args:    cobra.NoArgs,
		RunE:    servoCommand.RunServoList,
	})
	servoCmd.AddCommand(&cobra.Command{
		Use:   "Add",
		Short: "Add a Servo",
		Args:  cobra.MaximumNArgs(1),
		RunE:  servoCommand.RunAddServo,
	})
	servoCmd.AddCommand(&cobra.Command{
		Use:     "Remove",
		Aliases: []string{"rm"},
		Short:   "Remove a Servo",
		Args:    cobra.ExactArgs(1),
		RunE:    servoCommand.RunRemoveServo,
	})

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

var servos = []map[string]string{
	{
		"name": "opsani-dev",
		"host": "3.93.217.12",
		"port": "22",
		"user": "root",
		"path": "/root/dev.opsani.com/blake/oco",
	},
}

func (servoCmd *servoCommand) SSHAgent() ssh.AuthMethod {
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
	}
	return nil
}

func (servoCmd *servoCommand) RunAddServo(_ *cobra.Command, args []string) error {
	// Name, [user@]host[:port]/path
	return nil

	servos := []map[string]string{}
	var servo struct {
		Name string
		Host string
		Port string
		Path string
	}
	// var newServo servo

	if servo.Name == "" {
		servo.Name = args[0]
	}

	confirmed := false
	if servo.Name == "" {
		err := survey.AskOne(&survey.Input{
			Message: "Servo name?",
		}, &servo.Name, survey.WithValidator(survey.Required))
		if err != nil {
			return err
		}
		// prompt := &survey.Confirm{
		// 	Message: fmt.Sprintf("Write to %s?", opsani.ConfigFile),
		// }
		// cmd.AskOne(prompt, &servo.Name)
		// servos = append(servos, servo)
	}

	if confirmed {
		viper.Set("servos", servos)
		if err := viper.WriteConfig(); err != nil {
			return err
		}
	}

	return nil
}

func (servoCmd *servoCommand) RunRemoveServo(_ *cobra.Command, args []string) error {
	name := args[0]
	var servo map[string]string
	for _, s := range servos {
		if s["name"] == name {
			servo = s
			break
		}
	}
	if len(servo) == 0 {
		return fmt.Errorf("no such Servo %q", name)
	}

	confirmed := false
	if !confirmed {
		// prompt := &survey.Confirm{
		// 	Message: fmt.Sprintf("Remove Servo %q? from %q", servo["name"], opsani.ConfigFile),
		// }
		// servoCommand.AskOne(prompt, &confirmed)
	}

	if confirmed {
		if err := viper.WriteConfig(); err != nil {
			return err
		}
	}

	return nil
}

func (servoCmd *servoCommand) runInSSHSession(ctx context.Context, name string, runIt func(context.Context, map[string]string, *ssh.Session) error) error {

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

	var servo map[string]string
	for _, s := range servos {
		if s["name"] == name {
			servo = s
			break
		}
	}
	if len(servo) == 0 {
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
		User: servo["user"],
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
	client, err := ssh.Dial("tcp", servo["host"]+":"+servo["port"], config)
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

	return runIt(ctx, servo, session)
}

func (servoCmd *servoCommand) RunServoList(_ *cobra.Command, args []string) error {
	data := [][]string{}

	viper.Set("servos", servos)
	if err := viper.WriteConfigAs(servoCmd.ConfigFile); err != nil {
		return err
	}

	for _, servo := range servos {
		name := servo["name"]
		user := servo["user"]
		host := servo["host"]
		if port := servo["port"]; port != "" && port != "22" {
			host = host + ":" + port
		}
		path := "~/"
		if p := servo["path"]; p != "" {
			path = p
		}
		data = append(data, []string{name, user, host, path})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME", "USER", "HOST", "PATH"})
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
	table.AppendBulk(data) // Add Bulk Data
	table.Render()
	return nil
}

func (servoCmd *servoCommand) RunServoStatus(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	return servoCmd.runInSSHSession(ctx, args[0], func(ctx context.Context, servo map[string]string, session *ssh.Session) error {
		return servoCmd.runDockerComposeOverSSH("ps", nil, servo, session)
	})
}

func (servoCmd *servoCommand) RunServoStart(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	return servoCmd.runInSSHSession(ctx, args[0], func(ctx context.Context, servo map[string]string, session *ssh.Session) error {
		return servoCmd.runDockerComposeOverSSH("start", nil, servo, session)
	})
}

func (servoCmd *servoCommand) RunServoStop(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	return servoCmd.runInSSHSession(ctx, args[0], func(ctx context.Context, servo map[string]string, session *ssh.Session) error {
		return servoCmd.runDockerComposeOverSSH("stop", nil, servo, session)
	})
}

func (servoCmd *servoCommand) RunServoRestart(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	return servoCmd.runInSSHSession(ctx, args[0], func(ctx context.Context, servo map[string]string, session *ssh.Session) error {
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

func (servoCmd *servoCommand) runDockerComposeOverSSH(cmd string, args []string, servo map[string]string, session *ssh.Session) error {
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	if path := servo["path"]; path != "" {
		args = append(args, "cd", path+"&&")
	}
	args = append(args, "docker-compose", cmd)
	return session.Run(strings.Join(args, " "))
}

func (servoCmd *servoCommand) RunLogsSSHSession(ctx context.Context, servo map[string]string, session *ssh.Session) error {
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	args := []string{}
	if path := servo["path"]; path != "" {
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

func (servoCmd *servoCommand) RunShellOnSSHSession(ctx context.Context, servo map[string]string, session *ssh.Session) error {
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
