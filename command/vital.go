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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/charmbracelet/glamour"
	"github.com/fatih/color"
	"github.com/markbates/pkger"
	"github.com/mgutz/ansi"
	"github.com/mitchellh/go-homedir"
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
	markdown := `# Opsani Ignite

Ignite deploys a complete optimization experience onto your local workstation.

[Docker](https://www.docker.com/), [Kubernetes](https://kubernetes.io/), and [minikube](https://minikube.sigs.k8s.io/docs/) will be configured to run
a deployment of a simple web application, [Prometheus](https://prometheus.io/) for capturing metrics,
and a servo connected to your Opsani account.

Deployment will be done in a new minikube profile called **opsani-ignite** that is
isolated from your existing work.

Manifests generated during deployment are written to **./manifests**.`
	err := vitalCommand.DisplayMarkdown(markdown, false)
	if err != nil {
		return err
	}
	confirmed := false
	prompt := &survey.Confirm{
		Message: "Ready to get started?",
	}
	vitalCommand.AskOne(prompt, &confirmed)
	if !confirmed {
		return nil
	}
	fmt.Printf("\n💥 Let's do this thing.\n")

	bold := color.New(color.Bold).SprintFunc()
	vitalCommand.RunTaskWithSpinner(Task{
		Description: "checking for Docker runtime...",
		Success:     fmt.Sprintf("Docker %s found.", bold("{{.Version}}")),
		Failure:     "unable to find Docker",
		RunV: func() (interface{}, error) {
			cmd := exec.Command("docker", strings.Split("version --format v{{.Client.Version}}", " ")...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				return nil, err
			}
			return struct{ Version string }{Version: strings.TrimSpace(string(output))}, nil
		},
	})
	vitalCommand.RunTaskWithSpinner(Task{
		Description: "checking for Kubernetes...",
		Success:     fmt.Sprintf("Kubernetes %s found.", bold("{{ .clientVersion.gitVersion }}")),
		Failure:     "unable to find Kubernetes",
		RunV: func() (interface{}, error) {
			cmd := exec.Command("kubectl", strings.Split("version --client -o json", " ")...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				return nil, err
			}
			var versionInfo map[string]map[string]string
			err = json.Unmarshal(output, &versionInfo)
			if err != nil {
				return nil, err
			}
			return versionInfo, nil
		},
	})
	vitalCommand.RunTaskWithSpinner(Task{
		Description: "checking for minikube...",
		Success:     fmt.Sprintf("minikube %s found.", bold("{{ .minikubeVersion }}")),
		Failure:     "unable to find minikube",
		RunV: func() (interface{}, error) {
			cmd := exec.Command("minikube", strings.Split("version -o json", " ")...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				return nil, err
			}
			var versionInfo map[string]string
			err = json.Unmarshal(output, &versionInfo)
			if err != nil {
				return nil, err
			}
			return versionInfo, nil
		},
	})

	// Check to see if there is already an ignite cluster
	cmd := exec.Command("minikube", strings.Split("status -p opsani-ignite", " ")...)
	err = cmd.Run()
	if err == nil {
		recreate := false
		prompt := &survey.Confirm{
			Message: fmt.Sprintf(" There is an existing %q minikube profile. Do you want to recreate it?", "opsani-ignite"),
		}
		vitalCommand.AskOne(prompt, &recreate)
		if recreate {
			vitalCommand.RunTask(Task{
				Description: "deleting existing minikube profile...",
				Success:     fmt.Sprintf(`minikube profile %s deleted.`, bold("opsani-ignite")),
				Failure:     "failed deletion of minikube profile",
				RunW: func(w io.Writer) error {
					cmd := exec.Command("minikube", "delete", "-p", "opsani-ignite")
					cmd.Stdout = w
					cmd.Stderr = w
					cmd.Stdin = os.Stdin
					return cmd.Run()
				},
			})
		}
	}

	vitalCommand.RunTask(Task{
		Description: "creating a new minikube profile...",
		Success:     fmt.Sprintf(`minikube profile %s created.`, bold("opsani-ignite")),
		Failure:     "failed creation of minikube profile",
		RunW: func(w io.Writer) error {
			cmd := exec.Command("minikube", "start", "--memory=4096", "--cpus=4", "--wait=all", "-p", "opsani-ignite")
			cmd.Stdout = w
			cmd.Stderr = w
			cmd.Stdin = os.Stdin
			return cmd.Run()
		},
	})
	vitalCommand.RunTaskWithSpinner(Task{
		Description: "asking Opsani for an optimization engine...",
		Success:     "optimization engine acquired.",
		Failure:     "failed trying to acquire an optimization engine",
		Run: func() error {
			time.Sleep(4 * time.Second)
			return nil
		},
	})

	return vitalCommand.InstallKubernetesManifests(cobraCmd, args)
}

// DisplayMarkdown displays rendered Markdown in a pager
func (vitalCommand *vitalCommand) DisplayMarkdown(markdown string, paged bool) error {
	// Size paged output to the terminal
	fd := int(os.Stdin.Fd())
	termWidth, _, err := terminal.GetSize(fd)
	if err != nil {
		return err
	}
	if termWidth > 80 {
		termWidth = 80
	}

	r, err := glamour.NewTermRenderer(
		// TODO: detect background color and pick either the default dark or light theme
		glamour.WithStandardStyle("dark"),
		// wrap output at specific width
		glamour.WithWordWrap(termWidth),
	)
	if err != nil {
		return err
	}
	renderedMarkdown, err := r.Render(markdown)
	if err != nil {
		return err
	}

	// Let the user page lengthy content
	if paged {
		// Put terminal in interactive mode
		oldState, err := terminal.MakeRaw(fd)
		if err != nil {
			return err
		}
		defer terminal.Restore(fd, oldState)

		var pager io.WriteCloser
		cmd, pager := runPager()
		fmt.Fprint(pager, renderedMarkdown)
		pager.Close()
		return cmd.Wait()
	} else {
		fmt.Fprint(vitalCommand.OutOrStdout(), renderedMarkdown)
	}
	return nil
}

func (vitalCommand *vitalCommand) RunVital(cobraCmd *cobra.Command, args []string) error {
	markdown :=
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

	err := vitalCommand.DisplayMarkdown(markdown, true)
	if err != nil {
		return err
	}
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
	// ctx := context.Background()

	// cache escape codes and build strings manually
	// lime := ansi.ColorCode("green+h:black")
	blue := ansi.Blue
	reset := ansi.ColorCode("reset")
	whiteBold := ansi.ColorCode("white+b")
	// lightCyan := ansi.LightCyan

	// Pul the IMB image
	// imageRef := fmt.Sprintf("%s:%s", imbImageName, imbTargetVersion)
	// fmt.Printf("\n%s==>%s %sPulling %s...%s\n", blue, reset, whiteBold, imageRef, reset)
	// di, err := NewDockerInterface("")
	// if err != nil {
	//   return err
	// }
	//
	// err = di.PullImageWithProgressReporting(ctx, imageRef)
	// if err != nil {
	//   return err
	// }
	//
	// // Start discovery
	fmt.Printf("\n%s==>%s %sLaunching container...%s\n", blue, reset, whiteBold, reset)
	// return runIntelligentManifestBuilder("", imageRef)
	return nil
}

// TODO: This just duplicates exec.CombinedOutput
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
	if _, err := os.Stat("manifests"); os.IsNotExist(err) {
		e := os.Mkdir("manifests", 0755)
		if e != nil {
			return e
		}
	}

	bold := color.New(color.Bold).SprintFunc()
	err := pkger.Walk("/demo/manifests", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() || strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		// NOTE: The Prometheus manifests have custom resource definitions
		// That take awhile to propogate
		if info.Name() == "prometheus.yaml" {
			vitalCommand.RunTaskWithSpinner(Task{
				Description: "waiting for Prometheus custom resource definition to propogate...",
				Success:     "Prometheus custom resource definition is now available.",
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
			Description: fmt.Sprintf("applying manifest %s...", bold(info.Name())),
			Success:     fmt.Sprintf("manifest %s applied.", bold(info.Name())),
			Failure:     "manifest application failed",
			Run: func() error {
				f, err := pkger.Open(path)
				if err != nil {
					return err
				}

				manifestName := filepath.Base(path)
				manifestTemplate, err := ioutil.ReadAll(f)
				if err != nil {
					return err
				}

				tmpl, err := template.New("").Funcs(template.FuncMap{
					"base64encode": func(v string) string {
						return base64.StdEncoding.EncodeToString([]byte(v))
					},
				}).Parse(string(manifestTemplate))
				if err != nil {
					return err
				}

				cmd := exec.Command("kubectl", "--kubeconfig", pathToDefaultKubeconfig(), "apply", "--wait", "-f", "-")
				kubeCtlPipe, err := cmd.StdinPipe()
				if err != nil {
					return err
				}
				outputBuffer := new(bytes.Buffer)
				cmd.Stdout = outputBuffer
				cmd.Stderr = outputBuffer
				if err := cmd.Start(); err != nil {
					return fmt.Errorf("failed applying manifest %q: %w\n%s", manifestName, err, outputBuffer)
				}
				renderedManifest := new(bytes.Buffer)
				err = tmpl.Execute(renderedManifest, *vitalCommand.profile)
				if err != nil {
					panic(err)
				}
				fmt.Fprintln(kubeCtlPipe, renderedManifest)
				kubeCtlPipe.Close()
				if err := cmd.Wait(); err != nil {
					return fmt.Errorf("failed applying manifest %q: %w\n%s", manifestName, err, outputBuffer)
				}

				// Write the manifest
				manifestFile, err := os.Create(filepath.Join("manifests", info.Name()))
				if err != nil {
					return err
				}
				fmt.Fprintln(manifestFile, renderedManifest)
				manifestFile.Close()

				return nil
			}},
		)
	})
	if err != nil {
		return err
	}

	// Wait for Prometheus to become alive
	err = vitalCommand.RunTaskWithSpinner(Task{
		Description: "waiting for Prometheus pod...",
		Success:     "pod/prometheus-prometheus-0 is now running.",
		Failure:     "failed waiting for prometheus pod",
		Run: func() error {
			outcome := make(chan error)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			go func() {
				for {
					_, err := vitalCommand.run("kubectl", "wait", "--for", "condition=Ready", "pod/prometheus-prometheus-0")
					if err == nil {
						outcome <- nil
						return
					}
					select {
					case <-ctx.Done():
						outcome <- fmt.Errorf("failed waiting for Prometheus pod: %w", ctx.Err())
						return
					default:
						time.Sleep(1 * time.Second)
					}
				}
			}()
			select {
			case err := <-outcome:
				cancel()
				return err
			}
		},
	})
	if err != nil {
		return err
	}

	// Restart the servo so it can talk to Prometheus
	vitalCommand.run("kubectl", "rollout", "restart", "deployment", "servo")

	// Register a servo
	registry := NewServoRegistry(vitalCommand.viperCfg)
	if registry.ServoNamed("ignite") == nil {
		registry.AddServo(Servo{
			Name:       "ignite",
			Type:       "kubernetes",
			Namespace:  "default",
			Deployment: "servo",
		})
	}

	// Boom we are ready to roll
	boldBlue := color.New(color.FgHiBlue, color.Bold).SprintFunc()
	fmt.Fprintf(vitalCommand.OutOrStdout(), "\n🔥 %s\n", boldBlue("We have ignition"))
	fmt.Fprintf(vitalCommand.OutOrStdout(), "\nYour Servo is running in the %s deployment in Kubernetes\n", bold("servo"))
	fmt.Fprintf(vitalCommand.OutOrStdout(), "It has been registered as %s in the CLI\n", bold("ignite"))
	fmt.Fprintf(vitalCommand.OutOrStdout(),
		"\n%s  View servo commands: `%s`\n"+
			"%s  Follow servo logs: `%s`\n"+
			"%s  Watch pod status: `%s`\n"+
			"%s  Open Opsani console: `%s`\n\n",
		color.HiBlueString("ℹ"), color.YellowString("opsani servo --help"),
		color.HiBlueString("ℹ"), color.YellowString("opsani servo logs -f ignite"),
		color.HiBlueString("ℹ"), color.YellowString("kubectl get pods --watch"),
		color.HiBlueString("ℹ"), color.YellowString("opsani app console"))
	vitalCommand.Println(bold("Optimization results will begin reporting in the console shortly."))

	return err
}

func pathToDefaultKubeconfig() string {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return filepath.Join(home, ".kube", "config")
}
