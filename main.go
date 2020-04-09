package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"

	"github.com/stefaanc/golang-exec/runner"
	"github.com/stefaanc/golang-exec/runner/ssh"
	"github.com/stefaanc/golang-exec/script"
	"golang.org/x/crypto/ssh/terminal"
)

func main() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	username := flag.String("user", user.Username, "The user to authenticate as.")
	password := flag.String("password", "", "The password to authenticate with.")
	host := flag.String("host", "localhost", "The host to connect to.")
	sshMode := flag.Bool("ssh", false, "Execute via SSH.")
	flag.Parse()

	connectionType := "local"
	if *sshMode {
		connectionType = "ssh"

		if *password == "" {
			fmt.Print("Enter Password: ")
			bytePassword, err := terminal.ReadPassword(0)
			if err != nil {
				panic(err)
			}
			*password = string(bytePassword)
		}
	}

	c := ssh.Connection{
		Type:     connectionType,
		Host:     *host,
		Port:     22,
		User:     *username,
		Password: *password,
		Insecure: true,
	}

	err = runner.Run(&c, dockerScript, nil, os.Stdout, os.Stderr)
	if err != nil {
		var runnerErr runner.Error
		errors.As(err, &runnerErr)
		// fmt.Printf("exitcode: %d\n", runnerErr.ExitCode())
		log.Fatal(err)
	}
}

// TODO: Not sure why we need to cd ~/ and the --login flag on sh gets angry
var dockerScript = script.New("imb", "/bin/sh", "cd ~/ && docker run -it -v ~/.kube:/root/.kube imb")
