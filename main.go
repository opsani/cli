package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/docker/cli/cli/connhelper"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/kr/pretty"
	"github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/ssh/terminal"
)

const imbImageName string = "opsani/intelligent-manifest-builder"
const imbTargetVersion string = "latest" // TODO: Should be 1 for semantically versioned containers

/**
Pulls the latest Semantically Versioned tag of the manifest builder
*/
func pullManifestBuilderImage(ctx context.Context, cli *client.Client) {
	imageRef := fmt.Sprintf("%s:%s", imbImageName, imbTargetVersion)
	out, err := cli.ImagePull(ctx, imageRef, types.ImagePullOptions{})
	if err != nil {
		// panic(err)
		pretty.Printf("WARNING: Unable to pull Intelligent Manifest Builder image (%s): %s\n", imageRef, err.Error())
		return
	}

	defer out.Close()

	io.Copy(os.Stdout, out)
}

func runManifestBuilderContainer(ctx context.Context, cli *client.Client) {
	imageRef := fmt.Sprintf("%s:%s", imbImageName, imbTargetVersion)
	// FIXME: These paths are expanded locally but over an ssh transport
	// will resolve locally rather than on the remote host
	kubeDir, err := homedir.Expand("~/.kube")
	if err != nil {
		pretty.Println("No ~/.kube found")
		panic(err)
	}
	awsDir, err := homedir.Expand("~/.aws")
	if err != nil {
		pretty.Println("No ~/.aws found")
		panic(err)
	}
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        imageRef,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		OpenStdin:    true,
		StdinOnce:    false,
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:     mount.TypeBind,
				Source:   kubeDir,
				Target:   "/root/.kube",
				ReadOnly: true,
			},
			{
				Type:     mount.TypeBind,
				Source:   awsDir,
				Target:   "/root/.aws",
				ReadOnly: true,
			},
		},
	}, nil, "")
	if err != nil {
		panic(err)
	}

	// Start and attach to the container
	// The terminal/TTY hackery is necessary to enable interactive CLI applications
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

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
}

func main() {
	var dockerHost string
	flag.StringVar(&dockerHost, "host", "", "Specify the Docket host to connect to (overriding DOCKER_HOST)")
	flag.Parse()

	ctx := context.Background()
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
			return
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
		pretty.Println("Unable to create docker client")
		panic(err)
	}

	pullManifestBuilderImage(ctx, cli)
	runManifestBuilderContainer(ctx, cli)
}
