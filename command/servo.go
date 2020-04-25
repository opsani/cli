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

	"github.com/mitchellh/go-homedir"
	"github.com/prometheus/common/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
	"golang.org/x/crypto/ssh/terminal"
)

// NewServoCommand returns a new instance of the servo command
func NewServoCommand() *Command {
	logsCmd := NewCommandWithCobraCommand(&cobra.Command{
		Use:   "logs",
		Short: "View logs on a Servo",
		Args:  cobra.ExactArgs(1),
	}, func(cmd *Command) {
		cmd.RunE = RunServoLogs
	})

	servoCmd := NewCommandWithCobraCommand(&cobra.Command{
		Use:   "servo",
		Short: "Manage Servos",
		Args:  cobra.NoArgs,
	}, func(cmd *Command) {
		cmd.PersistentPreRunE = ReduceRunEFuncsO(InitConfigRunE, RequireConfigFileFlagToExistRunE, RequireInitRunE)
	})

	servoCmd.AddCommand(NewCommandWithCobraCommand(&cobra.Command{
		Use:   "ssh",
		Short: "SSH into a Servo",
		Args:  cobra.ExactArgs(1),
	}, func(cmd *Command) {
		cmd.RunE = RunServoSSH
	}).Command)

	servoCmd.AddCommand(logsCmd.Command)

	logsCmd.Flags().BoolP("follow", "f", false, "Follow log output")
	viper.BindPFlag("follow", logsCmd.Flags().Lookup("follow"))
	logsCmd.Flags().BoolP("timestamps", "t", false, "Show timestamps")
	viper.BindPFlag("timestamps", logsCmd.Flags().Lookup("timestamps"))
	logsCmd.Flags().StringP("lines", "l", "25", `Number of lines to show from the end of the logs (or "all").`)
	viper.BindPFlag("lines", logsCmd.Flags().Lookup("lines"))

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

func SSHAgent() ssh.AuthMethod {
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
	}
	return nil
}

func runInSSHSession(ctx context.Context, name string, runIt func(context.Context, map[string]string, *ssh.Session) error) error {

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
			SSHAgent(),
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

func RunServoLogs(cmd *Command, args []string) error {
	ctx := context.Background()
	return runInSSHSession(ctx, args[0], RunLogsSSHSession)
}

// RunConfig displays Opsani CLI config info
func RunServoSSH(cmd *Command, args []string) error {
	ctx := context.Background()
	return runInSSHSession(ctx, args[0], RunShellOnSSHSession)
}

func RunLogsSSHSession(ctx context.Context, servo map[string]string, session *ssh.Session) error {
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
	fmt.Printf("args: %v\n", args)
	return session.Run(strings.Join(args, " "))
}

func RunShellOnSSHSession(ctx context.Context, servo map[string]string, session *ssh.Session) error {
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
