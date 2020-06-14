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
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/charmbracelet/glamour"
	"github.com/fatih/color"
	"github.com/markbates/pkger"
	"github.com/mattn/go-colorable"
	"github.com/mgutz/ansi"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
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
func NewIgniteCommand(baseCmd *BaseCommand) *cobra.Command {
	vitalCommand := vitalCommand{BaseCommand: baseCmd}
	cobraCmd := &cobra.Command{
		Use:               "ignite",
		Short:             "Light up an interactive demo",
		Annotations:       map[string]string{"educational": "true"},
		Args:              cobra.NoArgs,
		PersistentPreRunE: nil,
		RunE:              vitalCommand.RunDemo,
	}

	loadGenCmd := &cobra.Command{
		Use:               "loadgen",
		Short:             "Learn about load generation in Opsani",
		Annotations:       map[string]string{"educational": "true"},
		Args:              cobra.NoArgs,
		PersistentPreRunE: nil,
		RunE:              vitalCommand.RunLearnLoadgen,
	}
	cobraCmd.AddCommand(loadGenCmd)
	adjustCmd := &cobra.Command{
		Use:               "adjust",
		Short:             "Learn about adjustments in Opsani",
		Annotations:       map[string]string{"educational": "true"},
		Args:              cobra.NoArgs,
		PersistentPreRunE: nil,
		RunE:              vitalCommand.RunLearnAdjust,
	}
	cobraCmd.AddCommand(adjustCmd)
	measureCmd := &cobra.Command{
		Use:               "measure",
		Short:             "Learn about measurements in Opsani",
		Annotations:       map[string]string{"educational": "true"},
		Args:              cobra.NoArgs,
		PersistentPreRunE: nil,
		RunE:              vitalCommand.RunLearnMeasure,
	}
	cobraCmd.AddCommand(measureCmd)

	bold := color.New(color.Bold).SprintFunc()

	startCmd := &cobra.Command{
		Use:               "start",
		Short:             "Start an Ignite cluster",
		Args:              cobra.NoArgs,
		PersistentPreRunE: nil,
		RunE: func(cmd *cobra.Command, args []string) error {
			mkCmd := exec.Command("minikube", "profile", "list", "-o", "json")
			output, err := mkCmd.Output()
			if err != nil {
				return err
			}
			result := gjson.GetBytes(output, `valid.#(Name=="opsani-ignite")`)
			if result.Exists() == false {
				return fmt.Errorf("minikube environment %q not found", "opsani-ignite")
			}

			return vitalCommand.RunTask(Task{
				Description: "starting minikube...",
				Success:     fmt.Sprintf(`minikube profile %s started.`, bold("opsani-ignite")),
				Failure:     "failed starting minikube",
				RunW: func(w io.Writer) error {
					cmd := exec.Command("minikube", "start", "-p", "opsani-ignite")
					cmd.Stdout = w
					cmd.Stderr = w
					cmd.Stdin = os.Stdin
					return cmd.Run()
				},
			})
		},
	}
	cobraCmd.AddCommand(startCmd)
	stopCmd := &cobra.Command{
		Use:               "stop",
		Short:             "Stop a running Ignite cluster",
		Args:              cobra.NoArgs,
		PersistentPreRunE: nil,
		RunE: func(cmd *cobra.Command, args []string) error {
			return vitalCommand.RunTask(Task{
				Description: "stopping minikube...",
				Success:     fmt.Sprintf(`minikube profile %s stopped.`, bold("opsani-ignite")),
				Failure:     "failed stopping minikube",
				RunW: func(w io.Writer) error {
					cmd := exec.Command("minikube", "stop", "-p", "opsani-ignite")
					cmd.Stdout = w
					cmd.Stderr = w
					cmd.Stdin = os.Stdin
					return cmd.Run()
				},
			})
		},
	}
	cobraCmd.AddCommand(stopCmd)
	statusCmd := &cobra.Command{
		Use:               "status",
		Short:             "Get the status of an Ignite cluster",
		Args:              cobra.NoArgs,
		PersistentPreRunE: nil,
		RunE: func(cmd *cobra.Command, args []string) error {
			return vitalCommand.RunTask(Task{
				Description: "getting minikube status...",
				Success:     fmt.Sprintf(`minikube profile %s status retrieved.`, bold("opsani-ignite")),
				Failure:     "failed getting minikube status",
				RunW: func(w io.Writer) error {
					cmd := exec.Command("minikube", "status", "-p", "opsani-ignite")
					cmd.Stdout = w
					cmd.Stderr = w
					cmd.Stdin = os.Stdin
					return cmd.Run()
				},
			})
		},
	}
	cobraCmd.AddCommand(statusCmd)
	deleteCmd := &cobra.Command{
		Use:               "delete",
		Short:             "Delete an Ignite cluster",
		Args:              cobra.NoArgs,
		PersistentPreRunE: nil,
		RunE: func(cmd *cobra.Command, args []string) error {
			return vitalCommand.RunTask(Task{
				Description: "deleting minikube profile...",
				Success:     fmt.Sprintf(`minikube profile %s deleted.`, bold("opsani-ignite")),
				Failure:     "failed deleting minikube profile",
				RunW: func(w io.Writer) error {
					cmd := exec.Command("minikube", "delete", "-p", "opsani-ignite")
					cmd.Stdout = w
					cmd.Stderr = w
					cmd.Stdin = os.Stdin
					return cmd.Run()
				},
			})
		},
	}
	cobraCmd.AddCommand(deleteCmd)

	return cobraCmd
}

func (vitalCommand *vitalCommand) RunLearnLoadgen(cobraCmd *cobra.Command, args []string) error {
	markdown := `# Opsani Ignite - Load Generation

Ignite is deployed with a load generation utility called [Vegeta](https://github.com/tsenart/vegeta).

It is a versatile, Golang based utility that supplies a constant request rate.

Vegeta is embedded within the servo assembly and is directly executed by the 
[servo-vegeta](https://github.com/opsani/servo-vegeta) connector to supply load during the *measure*
phase of a step.

The Vegeta configuration is part of the servo **config.yaml** file that is populated by
the **ConfigMap** defined in the **./manifests/servo-configmap.yaml**. The configuration
is nested under the **vegeta** key in the YAML.

To better understand the relationship between the load generation profile and how Opsani
evaluates performance and cost, try making changes to the **rate** key.

By default, the manifest is configured with a constant rate of **50/1s**, instructing Vegeta
to deliver 50 requests every second for the **duration** of the test.

Try increasing the rate to **500/1s** and applying the new manifest via:

` + "```console\nkubectl apply -f ./manifests/servo-configmap.yaml\nopsani servo restart\n```" + `

Then return to the Opsani Console and observe the differences in the next data points reported (~2 minutes later).`
	err := vitalCommand.DisplayMarkdown(markdown, true)
	if err != nil {
		return err
	}
	return nil
}

func (vitalCommand *vitalCommand) RunLearnAdjust(cobraCmd *cobra.Command, args []string) error {
	markdown := `# Opsani Ignite - Adjustments

Ignite deploys a servo onto Kubernetes alongside a simple web application called [co-http](https://github.com/opsani/co-http).

The servo is responsible for collecting measurements and making adjustments to the application under optimization.

Adjustments are deterministic changes applied to the resources that are supporting the application. In traditional
VM based cloud deployments, resources are mapped by instance types and families that define the amount of CPU,
memory, I/O & network bandwidth, etc.

Kubernetes deployments are interesting because the resourcing for a particular service or deployment is highly flexible
through dynamic allocations and distribution of the workload across replica sets.

When a Kubernetes based application, workload, or service is optimized by Opsani, parameters such as the CPU &
memory request and limit values will be explored by the optimizer to identify the optimal configuration at that moment
in time. Any optimizable resource in Opsani is modeled as a **component** and is explored by the optimizer in accordance
with **guard rails** defined during servo configuration. Components are described in terms of numeric **ranges** or 
**enumerations** that establish the bounds that the optimizer is permitted work in.

To better understand how adjustments are made and applied to a Kubernetes control plane, try making changes to
the **min**, **max**, and **step** values that are part of the **k8s/application/components** stanza in the the **ConfigMap** 
defined in the **./manifests/servo-configmap.yaml** file, applying the manifest, and restarting the servo deployment.

` + "```console\nkubectl apply -f ./manifests/servo-configmap.yaml\nopsani servo restart\n```" + `

Then return to the Opsani Console and observe the differences in the next data points reported (~2 minutes later).`
	err := vitalCommand.DisplayMarkdown(markdown, true)
	if err != nil {
		return err
	}
	return nil
}

func (vitalCommand *vitalCommand) RunLearnMeasure(cobraCmd *cobra.Command, args []string) error {
	markdown := `# Opsani Ignite - Measurements

Ignite deploys a servo onto Kubernetes alongside a simple web application called [co-http](https://github.com/opsani/co-http).

The servo is responsible for collecting measurements and making adjustments to the application under optimization.

Measurements are an aggregate report of metrics gathered from a source such as [Prometheus](https://prometheus.io/), 
[Datadog](https://www.datadoghq.com/), or [New Relic](https://newrelic.com/) during the reporting interval. Measurements are reported to the optimizer 
by the servo as a collection of time series values.

Measurements are critical to the optimization process because they provide the optimizer with data about how adjustments
impact the performance of the application under optimization. The optimizer evaluates the metrics reported against formulas 
that model the desired characteristics of the optimization solution. 

Nobody wishes to waste resources operating unnecessary infrastructure but applications are governed by service level objectives that define key performance indicators such
as acceptable error rates and latencies. This results in the defensive over-provisioning of infrastructure resources to provide
buffer against the unexpected. 

The optimizer is able to balance these concerns by applying data gathered from the application
under optimization "in the wild" to the configured optimization goals and make informed decisions about how to best adjust the
resources to achieve the desired results most efficiently.

Ignite gathers measurements from metrics gathered by Prometheus and reported by Vegeta during load generation.

To better understand how measurements impact the behavior of the optimizer, make changes to the servo configure as discussed
in the **opsani ignite loadgen** and **opsani ignite adjust** commands. The load profile and resources allocated to the application 
have a direct impact on application performance that is immediately seen in the data points reported to the Opsani console.
`
	err := vitalCommand.DisplayMarkdown(markdown, true)
	if err != nil {
		return err
	}
	return nil
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
	fmt.Fprintf(vitalCommand.OutOrStdout(), "\n💥 Let's do this thing.\n")

	bold := color.New(color.Bold).SprintFunc()
	err = vitalCommand.RunTaskWithSpinner(Task{
		Description: "checking for Docker runtime...",
		Success:     fmt.Sprintf("Docker %s found.", bold("{{.Version}}")),
		Failure:     "unable to find Docker",
		RunV: func() (interface{}, error) {
			path, err := exec.LookPath("docker")
			if err != nil {
				return nil, fmt.Errorf("docker not found on path")
			}
			cmd := exec.Command(path, strings.Split("version --format v{{.Client.Version}}", " ")...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				return nil, fmt.Errorf("failed retrieving Docker version: %w: %s", err, output)
			}
			return struct{ Version string }{Version: strings.TrimSpace(string(output))}, nil
		},
	})
	if err != nil {
		return err
	}

	err = vitalCommand.RunTaskWithSpinner(Task{
		Description: "checking for Kubernetes...",
		Success:     fmt.Sprintf("Kubernetes %s found.", bold("{{ .clientVersion.gitVersion }}")),
		Failure:     "unable to find Kubernetes",
		RunV: func() (interface{}, error) {
			path, err := exec.LookPath("kubectl")
			if err != nil {
				return nil, fmt.Errorf("kubectl not found on path")
			}
			cmd := exec.Command(path, strings.Split("version --client -o json", " ")...)
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
	if err != nil {
		return err
	}

	err = vitalCommand.RunTaskWithSpinner(Task{
		Description: "checking for minikube...",
		Success:     fmt.Sprintf("minikube %s found.", bold("{{ .minikubeVersion }}")),
		Failure:     "unable to find minikube",
		RunV: func() (interface{}, error) {
			path, err := exec.LookPath("minikube")
			if err != nil {
				return nil, fmt.Errorf("minikube not found on path")
			}
			cmd := exec.Command(path, strings.Split("version -o json", " ")...)
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
	if err != nil {
		return err
	}

	// Check to see if there is already an ignite cluster
	existingProfile := false
	mkCmd := exec.Command("minikube", "profile", "list", "-o", "json")
	output, err := mkCmd.Output()
	if err == nil {
		result := gjson.GetBytes(output, `valid.#(Name=="opsani-ignite")`)
		existingProfile = result.Exists()
	} else {
		results := gjson.GetManyBytes(output, "error.Op", "error.Err")
		if results[0].String() == "open" && results[1].Int() == 2 {
			// Ignore -- this means there aren't any profiles
		} else {
			return fmt.Errorf("failed listing minikube profiles: %w: %s", err, output)
		}
	}
	if existingProfile {
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

	err = vitalCommand.RunTask(Task{
		Description: "creating a new minikube profile...",
		Success:     fmt.Sprintf(`minikube profile %s created.`, bold("opsani-ignite")),
		Failure:     "failed creation of minikube profile",
		RunW: func(w io.Writer) error {
			cmd := exec.Command("minikube", "start", "--memory=4096", "--cpus=4", "--wait=all", "-p", "opsani-ignite")
			if runtime.GOOS == "windows" {
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
			} else {
				cmd.Stdout = w
				cmd.Stderr = w
			}
			cmd.Stdin = os.Stdin
			return cmd.Run()
		},
	})
	if err != nil {
		return err
	}

	err = vitalCommand.RunTaskWithSpinner(Task{
		Description: "asking Opsani for an optimization engine...",
		Success:     "optimization engine acquired.",
		Failure:     "failed trying to acquire an optimization engine",
		Run: func() error {
			time.Sleep(4 * time.Second)
			return nil
		},
	})
	if err != nil {
		return err
	}

	return vitalCommand.InstallKubernetesManifests(cobraCmd, args)
}

// DisplayMarkdown displays rendered Markdown in a pager
func (vitalCommand *vitalCommand) DisplayMarkdown(markdown string, paged bool) error {
	fd := int(os.Stdin.Fd())
	r, err := glamour.NewTermRenderer(
		// TODO: detect background color and pick either the default dark or light theme
		glamour.WithStandardStyle("dark"),
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
		cmd, pager, err := runPager()
		if err != nil {
			return err
		}
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

func runPager() (*exec.Cmd, io.WriteCloser, error) {
	var cmd *exec.Cmd
	// if runtime.GOOS == "windows" {
	path, err := exec.LookPath("less")
	if err == nil {
		cmd = exec.Command(path, ArgsS("-F -g -i -M -R -S -w -X -z-4")...)
	} else {
		pager := os.Getenv("PAGER")
		if pager == "" {
			pager = "more"
		}
		path, err = exec.LookPath(pager)
		if err != nil {
			return nil, nil, err
		}
		cmd = exec.Command(path)
	}

	// } else {
	// 	cmd
	// }

	// cmd := exec.Command("powershell.exe", "-Command", "& {Out-Host -Paging -}") //"powershell", "{Out-Host", "-Paging}")
	out, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, err
	}
	cmd.Stdout = colorable.NewColorableStdout()
	cmd.Stderr = colorable.NewColorableStderr()
	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}
	return cmd, out, err
}

func (vitalCommand *vitalCommand) RunVitalDiscovery(cobraCmd *cobra.Command, args []string) error {
	// ctx := context.Background()

	// cache escape codes and build strings manually
	// lime := ansi.ColorCode("green+h:black")
	blue := ansi.Blue
	reset := ansi.ColorCode("reset")
	whiteBold := ansi.ColorCode("white+b")
	// lightCyan := ansi.LightCyan

	// Pull the IMB image
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

	// Apply the desired backend configuration
	err = vitalCommand.RunTaskWithSpinner(Task{
		Description: "configuring optimizer for ignite...",
		Success:     "optimizer configured.",
		Failure:     "failed configuring optimizer for ignite",
		Run: func() error {
			client := vitalCommand.NewAPIClient()
			body, err := json.MarshalIndent(map[string]map[string]string{
				"optimization": {
					"perf": "latency_90th",
				},
			}, "", "  ")
			if err != nil {
				return err
			}

			_, err = client.PatchConfigFromBody(body, true)
			if err != nil {
				return err
			}
			return nil
		},
	})
	if err != nil {
		return err
	}

	// Restart the servo so it can talk to Prometheus
	vitalCommand.run("kubectl", "rollout", "restart", "deployment", "servo")

	// Attach the servo
	attachServo := (vitalCommand.profile.Servo == (Servo{}))
	if !attachServo {
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Existing servo attached to %q. Overwrite?", vitalCommand.profile.Name),
		}
		vitalCommand.AskOne(prompt, &attachServo)
	}
	if attachServo {
		registry, err := NewProfileRegistry(vitalCommand.viperCfg)
		if err != nil {
			return err
		}
		profile := registry.ProfileNamed(vitalCommand.profile.Name)
		profile.Servo = Servo{
			Type:       "kubernetes",
			Namespace:  "default",
			Deployment: "servo",
		}
		if err = registry.Save(); err != nil {
			return err
		}
	}

	profileOption := ""
	if !vitalCommand.profile.IsActive() {
		profileOption = fmt.Sprintf("-p %s ", vitalCommand.profile.Name)
	}

	// Boom we are ready to roll
	boldBlue := color.New(color.FgHiBlue, color.Bold).SprintFunc()
	fmt.Fprintf(vitalCommand.OutOrStdout(), "\n🔥 %s\n", boldBlue("We have ignition"))
	fmt.Fprintf(vitalCommand.OutOrStdout(), "\n%s  Servo running in Kubernetes %s\n", color.HiBlueString("ℹ"), bold("deployments/servo"))
	fmt.Fprintf(vitalCommand.OutOrStdout(), "%s  Servo attached to opsani profile %s\n", color.HiBlueString("ℹ"), bold(vitalCommand.profile.Name))
	fmt.Fprintf(vitalCommand.OutOrStdout(), "%s  Manifests written to %s\n", color.HiBlueString("ℹ"), bold("./manifests"))
	fmt.Fprintf(vitalCommand.OutOrStdout(),
		"\n%s  View ignite subcommands: `%s`\n"+
			"%s  View servo subcommands: `%s`\n"+
			"%s  Follow servo logs: `%s`\n"+
			"%s  Watch pod status: `%s`\n"+
			"%s  Open Opsani console: `%s`\n\n",
		color.HiGreenString("❯"), color.YellowString(fmt.Sprintf("opsani %signite --help", profileOption)),
		color.HiGreenString("❯"), color.YellowString(fmt.Sprintf("opsani %sservo --help", profileOption)),
		color.HiGreenString("❯"), color.YellowString(fmt.Sprintf("opsani %sservo logs -f", profileOption)),
		color.HiGreenString("❯"), color.YellowString("kubectl get pods --watch"),
		color.HiGreenString("❯"), color.YellowString(fmt.Sprintf("opsani %sconsole", profileOption)))
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
