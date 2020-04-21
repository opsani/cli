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
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"context"
	"io"

	"github.com/docker/cli/cli/connhelper"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"golang.org/x/crypto/ssh/terminal"
)

// DockerInterface provides utilities and UI affordances for working with Docker in the CLI
type DockerInterface struct {
	dockerClient *client.Client
}

// NewDockerInterface initializes and returns a new DockerInterface for interacting with Docker
func NewDockerInterface(dockerHost string) (*DockerInterface, error) {
	var clientOpts []client.Opt
	clientOpts = append(clientOpts,
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)

	// Resolve the Docker host using the connection helpers
	// This supports the resolution of ssh:// URL schemes for tunneled execution
	if dockerHost != "" {
		helper, err := connhelper.GetConnectionHelper(dockerHost)
		if err != nil {
			return nil, err
		}

		httpClient := &http.Client{
			// No tls
			// No proxy
			Transport: &http.Transport{
				DialContext: helper.Dialer,
			},
		}

		clientOpts = append(clientOpts,
			client.WithHTTPClient(httpClient),
			client.WithHost(helper.Host),
			client.WithDialContext(helper.Dialer),
		)
	}

	cli, err := client.NewClientWithOpts(clientOpts...)
	if err != nil {
		return nil, err
	}
	return NewDockerInterfaceWithDockerClient(cli)
}

// NewDockerInterfaceWithDockerClient initializes and returns a new DockerInterface for interacting with Docker using the given Docker client
func NewDockerInterfaceWithDockerClient(dockerClient *client.Client) (*DockerInterface, error) {
	return &DockerInterface{
		dockerClient: dockerClient,
	}, nil
}

func (di *DockerInterface) DockerClient() *client.Client {
	return di.dockerClient
}

// Struct representing events returned from image pulling
type pullEvent struct {
	ID             string `json:"id"`
	Status         string `json:"status"`
	Error          string `json:"error,omitempty"`
	Progress       string `json:"progress,omitempty"`
	ProgressDetail struct {
		Current int `json:"current"`
		Total   int `json:"total"`
	} `json:"progressDetail"`
}

// Cursor structure that implements some methods
// for manipulating command line's cursor
type Cursor struct{}

func (cursor *Cursor) hide() {
	fmt.Printf("\033[?25l")
}

func (cursor *Cursor) show() {
	fmt.Printf("\033[?25h")
}

func (cursor *Cursor) moveUp(rows int) {
	fmt.Printf("\033[%dF", rows)
}

func (cursor *Cursor) moveDown(rows int) {
	fmt.Printf("\033[%dE", rows)
}

func (cursor *Cursor) clearLine() {
	fmt.Printf("\033[2K")
}

// PullImageWithProgressReporting pulls a Docker image from a repository and outputs progress to the terminal
func (di *DockerInterface) PullImageWithProgressReporting(ctx context.Context, imageRef string) error {
	out, err := di.dockerClient.ImagePull(ctx, imageRef, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer out.Close()

	cursor := Cursor{}
	layers := make([]string, 0)
	oldIndex := len(layers)

	var event *pullEvent
	decoder := json.NewDecoder(out)

	fmt.Printf("\n")
	cursor.hide()

	for {
		if err := decoder.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		imageID := event.ID

		// Check if the line is one of the final two ones
		if strings.HasPrefix(event.Status, "Digest:") || strings.HasPrefix(event.Status, "Status:") {
			fmt.Printf("%s\n", event.Status)
			continue
		}

		// Check if ID has already passed once
		index := 0
		for i, v := range layers {
			if v == imageID {
				index = i + 1
				break
			}
		}

		// Move the cursor
		if index > 0 {
			diff := index - oldIndex

			if diff > 1 {
				down := diff - 1
				cursor.moveDown(down)
			} else if diff < 1 {
				up := diff*(-1) + 1
				cursor.moveUp(up)
			}

			oldIndex = index
		} else {
			layers = append(layers, event.ID)
			diff := len(layers) - oldIndex

			if diff > 1 {
				cursor.moveDown(diff) // Return to the last row
			}

			oldIndex = len(layers)
		}

		cursor.clearLine()

		if event.Status == "Pull complete" {
			fmt.Printf("%s: %s\n", event.ID, event.Status)
		} else {
			fmt.Printf("%s: %s %s\n", event.ID, event.Status, event.Progress)
		}

	}

	cursor.show()
	return nil
}

/**
InteractiveContainerConfig describes the configuration of an interactive container execution
*/
type InteractiveContainerConfig struct {
	ContainerConfig        *container.Config
	ContainerStartOptions  types.ContainerStartOptions
	ContainerAttachOptions types.ContainerAttachOptions
	HostConfig             *container.HostConfig
	CompletionCallback     func(context.Context, *DockerInterface, container.ContainerCreateCreatedBody, container.ContainerWaitOKBody)
}

// NewInteractiveContainerConfigWithImageRef returns a new config object for running a Docker container interactively
func NewInteractiveContainerConfigWithImageRef(imageRef string) *InteractiveContainerConfig {
	return &InteractiveContainerConfig{
		ContainerConfig: &container.Config{
			Image:        imageRef,
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          true,
			OpenStdin:    true,
			StdinOnce:    false,
		},
		HostConfig:            &container.HostConfig{},
		ContainerStartOptions: types.ContainerStartOptions{},
		ContainerAttachOptions: types.ContainerAttachOptions{
			Stderr: true,
			Stdout: true,
			Stdin:  true,
			Stream: true,
		},
	}
}

// RunInteractiveContainer starts and runs a container interactively, allowing user interaction with console applications
func (di *DockerInterface) RunInteractiveContainer(ctx context.Context, icc *InteractiveContainerConfig) error {
	resp, err := di.dockerClient.ContainerCreate(ctx, icc.ContainerConfig, icc.HostConfig, nil, "")
	if err != nil {
		return err
	}

	// Start and attach to the container
	// The terminal/TTY hackery is necessary to enable interactive CLI applications
	if err := di.dockerClient.ContainerStart(ctx, resp.ID, icc.ContainerStartOptions); err != nil {
		return err
	}

	fd := int(os.Stdin.Fd())
	oldState, err := terminal.MakeRaw(fd)
	if err != nil {
		return err
	}
	defer terminal.Restore(fd, oldState)

	termWidth, termHeight, err := terminal.GetSize(fd)
	if err != nil {
		return err
	}

	if err = di.dockerClient.ContainerResize(ctx, resp.ID, types.ResizeOptions{Width: uint(termWidth), Height: uint(termHeight)}); err != nil {
		return err
	}

	waiter, err := di.dockerClient.ContainerAttach(ctx, resp.ID, icc.ContainerAttachOptions)
	go io.Copy(os.Stdout, waiter.Reader)
	go io.Copy(waiter.Conn, os.Stdin)

	statusCh, errCh := di.dockerClient.ContainerWait(ctx, resp.ID, container.WaitConditionNextExit)
	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case s := <-statusCh:
		if icc.CompletionCallback != nil {
			icc.CompletionCallback(ctx, di, resp, s)
		}
	}

	out, err := di.dockerClient.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		return err
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	return nil
}
