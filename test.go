package main

import (
	"context"
	"io"
	"net/http"
	"os"

	"github.com/docker/cli/cli/connhelper"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/kr/pretty"
	"golang.org/x/crypto/ssh/terminal"
)

var inout chan []byte

func main() {
	ctx := context.Background()
	helper, err := connhelper.GetConnectionHelper("ssh://localhost")

	if err != nil {
		return
	}

	httpClient := &http.Client{
		// No tls
		// No proxy
		Transport: &http.Transport{
			DialContext: helper.Dialer,
		},
	}

	var clientOpts []client.Opt

	clientOpts = append(clientOpts,
		client.WithHTTPClient(httpClient),
		client.WithHost(helper.Host),
		client.WithDialContext(helper.Dialer),
	)

	version := os.Getenv("DOCKER_API_VERSION")

	if version != "" {
		clientOpts = append(clientOpts, client.WithVersion(version))
	} else {
		clientOpts = append(clientOpts, client.WithAPIVersionNegotiation())
	}

	cli, err := client.NewClientWithOpts(clientOpts...)

	if err != nil {
		pretty.Println("Unable to create docker client")
		panic(err)
	}

	// pretty.Println(cl.ImageList(context.Background(), types.ImageListOptions{}))
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        "imb",
		AttachStderr: true,
		AttachStdin:  true,
		Tty:          true,
		AttachStdout: true,
		OpenStdin:    true,
		StdinOnce:    false,
		Cmd:          os.Args[1:],
	}, &container.HostConfig{

		Mounts: []mount.Mount{
			mount.Mount{
				Type:   mount.TypeBind,
				Source: "/Users/blakewatters/.kube",
				Target: "/root/.kube",
			},
		},
	}, nil, "")
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	fd := int(os.Stdin.Fd())
	oldState, err := terminal.MakeRaw(fd)
	if err != nil {
		panic(err)
	}

	termWidth, termHeight, err := terminal.GetSize(fd)
	if err != nil {
		panic(err)
	}

	if err = cli.ContainerResize(ctx, resp.ID, types.ResizeOptions{Width: uint(termWidth), Height: uint(termHeight)}); err != nil {
		panic(err)
	}

	waiter, err := cli.ContainerAttach(ctx, resp.ID, types.ContainerAttachOptions{
		Stderr: true,
		Stdout: true,
		Stdin:  true,
		Stream: true,
	})
	defer cli.Close()
	go io.Copy(os.Stdout, waiter.Reader)
	go io.Copy(waiter.Conn, os.Stdin)

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}
	terminal.Restore(fd, oldState)
}
