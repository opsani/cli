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
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/AlecAivazis/survey/v2"
	"github.com/charmbracelet/glamour"
	"github.com/mgutz/ansi"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

type vitalCommand struct {
	*BaseCommand
}

// NewVitalCommand returns a new instance of the vital command
func NewVitalCommand(baseCmd *BaseCommand) *cobra.Command {
	vitalCommand := vitalCommand{BaseCommand: baseCmd}
	cobraCmd := &cobra.Command{
		Use:               "vital",
		Short:             "Start optimizing",
		Args:              cobra.NoArgs,
		PersistentPreRunE: nil,
		RunE:              vitalCommand.RunVital,
	}

	return cobraCmd
}

func (vitalCommand *vitalCommand) RunVital(cobraCmd *cobra.Command, args []string) error {
	in :=
		`# Opsani Vital

## Let's talk about your cloud costs

It's the worst kept secret in modern tech. We're all spending way more on infrastructure than is necesary.

But it's not our fault. Our applications have become too big and complicated to optimize.

Until now.

## Better living through machine learning...

Opsani utilizes state of the art machine learning technology to continuously optimize your applications for *cost* and *performance*.

## Getting started

To start optimizing, a Servo must be deployed into your environment.

A Servo is a lightweight container that lets Opsani know what is going on in your application and recommend optimizations.

This app is designed to assist you in assembling and deploying a Servo through the miracle of automation and sensible defaults.

The process looks like...

- [x] Register for Vital
- [x] Install Opsani
- [x] Read this doc
- [ ] Deploy the Servo
- [ ] Start optimizing

## Things to keep in mind

All software run and deployed is Open Source. Opsani supports manual and assisted integrations if you like to do things the hard way.

Over the next 20 minutes, we will gather details about your application, the deployment environment, and your optimization goals.

The process will involve cloning Git repositories, connecting to your metrics & orchestration systems, and running Docker containers.

As tasks are completed, artifacts will be generated and saved onto this workstation.

Everything is logged, you can be pause and resume at any time, and important items will require confirmation.

Once this is wrapped up, you can start optimizing immediately.`

	// Size paged output to the terminal
	fd := int(os.Stdin.Fd())
	oldState, err := terminal.MakeRaw(fd)
	if err != nil {
		return err
	}
	defer terminal.Restore(fd, oldState)

	termWidth, _, err := terminal.GetSize(fd)
	if err != nil {
		return err
	}

	r, _ := glamour.NewTermRenderer(
		// detect background color and pick either the default dark or light theme
		glamour.WithStandardStyle("dark"),
		// wrap output at specific width
		glamour.WithWordWrap(termWidth),
	)
	out, _ := r.Render(in)

	var pager io.WriteCloser
	cmd, pager := runPager()
	fmt.Fprintln(pager, out)

	// Let the user page
	pager.Close()
	err = cmd.Wait()
	if err != nil {
		return err
	}

	// Let's get on with it!
	confirmed := false
	prompt := &survey.Confirm{
		Message: "Ready to get started?",
	}
	vitalCommand.AskOne(prompt, &confirmed)
	if confirmed {
		fmt.Printf("\n💥 Let's do this thing.\n")
		return vitalCommand.RunVitalDiscovery(cobraCmd, args)
	}

	return nil
}

func runPager() (*exec.Cmd, io.WriteCloser) {
	pager := os.Getenv("PAGER")
	if pager == "" {
		pager = "more"
	}
	cmd := exec.Command(pager)
	out, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	return cmd, out
}

func (vitalCommand *vitalCommand) RunVitalDiscovery(cobraCmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// cache escape codes and build strings manually
	// lime := ansi.ColorCode("green+h:black")
	blue := ansi.Blue
	reset := ansi.ColorCode("reset")
	whiteBold := ansi.ColorCode("white+b")
	// lightCyan := ansi.LightCyan

	// Pul the IMB image
	imageRef := fmt.Sprintf("%s:%s", imbImageName, imbTargetVersion)
	fmt.Printf("\n%s==>%s %sPulling %s...%s\n", blue, reset, whiteBold, imageRef, reset)
	di, err := NewDockerInterface("")
	if err != nil {
		return err
	}

	err = di.PullImageWithProgressReporting(ctx, imageRef)
	if err != nil {
		return err
	}

	// Start discovery
	fmt.Printf("\n%s==>%s %sLaunching container...%s\n", blue, reset, whiteBold, reset)
	return runIntelligentManifestBuilder("", imageRef)
}
