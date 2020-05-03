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
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/charmbracelet/glamour"
	"github.com/fatih/color"
	"github.com/markbates/pkger"
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

// NewDemoCommand returns a new instance of the demo command
func NewDemoCommand(baseCmd *BaseCommand) *cobra.Command {
	vitalCommand := vitalCommand{BaseCommand: baseCmd}
	cobraCmd := &cobra.Command{
		Use:               "ignite",
		Short:             "Light up an interactive demo",
		Args:              cobra.NoArgs,
		PersistentPreRunE: nil,
		RunE:              vitalCommand.RunDemo,
	}

	return cobraCmd
}

func (vitalCommand *vitalCommand) RunDemo(cobraCmd *cobra.Command, args []string) error {
	// TODO: Check for Docker
	// Check for Minikube
	// Check for existing instance and prompt to reset it
	// Fake connecting to the Opsani backend and probing the environment
	// Light up Minikube
	// Open the console
	fmt.Println("We are about to deploy a web app and a Servo in a local Kubernetes cluster.")
	fmt.Println("Everything is isolated from your existing work and available to you under an Open Source license.")
	confirmed := false
	prompt := &survey.Confirm{
		Message: "Ready to get started?",
	}
	vitalCommand.AskOne(prompt, &confirmed)
	if !confirmed {
		return nil
	}
	fmt.Printf("\n💥 Let's do this thing.\n")

	vitalCommand.RunTaskWithSpinner(Task{
		Description: "checking for Docker runtime...",
		Success:     "Docker v19.03.8 found.",
		Failure:     "optimization engine deployment failed",
		Run: func() error {
			time.Sleep(2 * time.Second)
			return nil
		},
	})
	vitalCommand.RunTaskWithSpinner(Task{
		Description: "checking for Kubernetes...",
		Success:     "Kubernetes v1.18.0 found.",
		Failure:     "optimization engine deployment failed",
		Run: func() error {
			time.Sleep(1 * time.Second)
			return nil
		},
	})
	vitalCommand.RunTaskWithSpinner(Task{
		Description: "checking for minikube...",
		Success:     "minikube v1.9.2 found.",
		Failure:     "optimization engine deployment failed",
		Run: func() error {
			time.Sleep(1 * time.Second)
			return nil
		},
	})
	vitalCommand.RunTaskWithSpinner(Task{
		Description: "creating a new minikube profile...",
		Success:     `minikube profile "opsani-demo" created.`,
		Failure:     "optimization engine deployment failed",
		Run: func() error {
			time.Sleep(1 * time.Second)
			return nil
		},
	})
	vitalCommand.RunTaskWithSpinner(Task{
		Description: "asking Opsani for an optimization engine...",
		Success:     "optimization engine acquired.",
		Failure:     "optimization engine deployment failed",
		Run: func() error {
			time.Sleep(4 * time.Second)
			return nil
		},
	})

	return vitalCommand.InstallKubernetesManifests(cobraCmd, args)
}

func (vitalCommand *vitalCommand) RunVital(cobraCmd *cobra.Command, args []string) error {
	in :=
		`# Opsani Vital

## Let's talk about your cloud costs

It's the worst kept secret in tech. We're all spending way more on infrastructure than is necessary.

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
	terminal.Restore(fd, oldState)
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

func (vitalCommand *vitalCommand) run(name string, args ...string) (*bytes.Buffer, error) {
	outputBuffer := new(bytes.Buffer)
	cmd := exec.Command(name, args...)
	cmd.Stdout = outputBuffer
	cmd.Stderr = outputBuffer
	err := cmd.Run()
	return outputBuffer, err
}

func init() {
	pkger.Include("/demo/manifests")
}

func (vitalCommand *vitalCommand) InstallKubernetesManifests(cobraCmd *cobra.Command, args []string) error {
	err := vitalCommand.RunTaskWithSpinner(Task{
		Description: "creating \"opsani-servo\" namespace...",
		Success:     "created \"opsani-servo\" namespace.",
		Failure:     "namespace creation failed.\n",
		Run: func() error {
			output, err := vitalCommand.run("kubectl", "create", "namespace", "opsani-servo")
			if err != nil {
				return fmt.Errorf("%s: %w", output, err)
			}
			return nil
		},
	})
	if err != nil {
		return err
	}

	err = vitalCommand.RunTaskWithSpinner(Task{
		Description: "creating secrets...",
		Success:     "your secrets are safe and sound.",
		Failure:     "secret creation failed.\n",
		Run: func() error {
			output, err := vitalCommand.run("kubectl", "--kubeconfig", pathToDefaultKubeconfig(), "create", "secret", "generic", "opsani-servo-auth",
				"--from-literal", fmt.Sprintf("token=%s", vitalCommand.AccessToken()), "--namespace", "opsani-servo")
			if err != nil {
				return fmt.Errorf("%s: %w", output, err)
			}
			return nil
		},
	})
	if err != nil {
		return err
	}

	org, app := vitalCommand.AppComponents()
	vars := struct {
		Account string
		App     string
	}{
		Account: org,
		App:     app,
	}
	err = pkger.Walk("/demo/manifests", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() || strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		// NOTE: The Prometheus manifests have custom resource definitions
		// That take awhile to propogate
		if info.Name() == "prometheus.yaml" {
			vitalCommand.RunTaskWithSpinner(Task{
				Description: "waiting for prometheus custom resource definition to propogate...",
				Success:     "prometheus custom resource definition is now available.",
				Run: func() error {
					for {
						c := exec.Command("kubectl", "get", "prometheuses")
						err = c.Run()
						if err == nil {
							break
						}
						// Keep waiting
						time.Sleep(2 * time.Second)
					}
					return nil
				},
			})
		}

		return vitalCommand.RunTaskWithSpinner(Task{
			Description: fmt.Sprintf("applying manifest %q...", info.Name()),
			Success:     fmt.Sprintf("manifest %q applied.", info.Name()),
			Failure:     "FAILED!",
			Run: func() error {
				f, err := pkger.Open(path)
				if err != nil {
					return err
				}

				manifestTemplate, err := ioutil.ReadAll(f)
				if err != nil {
					return err
				}

				tmpl, err := template.New("").Parse(string(manifestTemplate))
				if err != nil {
					return err
				}

				cmd := exec.Command("kubectl", "--kubeconfig", pathToDefaultKubeconfig(), "apply", "--wait", "-f", "-")
				out, err := cmd.StdinPipe()
				if err != nil {
					return err
				}
				outputBuffer := new(bytes.Buffer)
				cmd.Stdout = outputBuffer
				cmd.Stderr = outputBuffer
				if err := cmd.Start(); err != nil {
					return fmt.Errorf("%s: %w", outputBuffer, err)
				}
				buf := new(bytes.Buffer)
				err = tmpl.Execute(buf, vars)
				if err != nil {
					panic(err)
				}
				fmt.Fprintln(out, buf)
				out.Close()
				if err := cmd.Wait(); err != nil {
					return fmt.Errorf("%s: %w", buf, err)
				}

				return nil
			}},
		)
	})

	// Restart the Servo deployment in case Prometheus DNS was live
	vitalCommand.run("kubectl", "rollout", "restart", "deployment", "opsani-servo", "-n", "opsani-servo")

	// Boom we are ready to roll
	c := color.New(color.FgBlue, color.Bold).SprintFunc()
	fmt.Fprintf(vitalCommand.OutOrStdout(), "🔥  %s\n", c("Deployment completed."))
	fmt.Fprintf(vitalCommand.OutOrStdout(), "\n\nThe local install of minikube is now running a Servo assembly\n"+
		"that is connected to your Opsani optimization engine.\nYou can observe the logs by executing `kubectl get pods -A`\n"+
		"and then issuing a `kubectl logs -f` command against the Servo pod.\n\n")

	confirmed := false
	prompt := &survey.Confirm{
		Message: "The Servo will begin reporting results to the Opsani Console shortly.\nWould you like to go there now?",
	}
	vitalCommand.AskOne(prompt, &confirmed)
	if confirmed {
		org, appID := vitalCommand.GetAppComponents()
		url := fmt.Sprintf("https://console.opsani.com/accounts/%s/applications/%s", org, appID)
		openURLInDefaultBrowser(url)
	}

	return err
}
