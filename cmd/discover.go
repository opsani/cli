/*
Copyright © 2020 Blake Watters <blake@opsani.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/docker/cli/cli/connhelper"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/kr/pretty"
	"github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/ssh/terminal"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const imbImageName string = "opsani/k8s-imb"
const imbTargetVersion string = "latest" // TODO: Should be 1 for semantically versioned containers

// Configuration options bound via Cobra
var discoverConfig = struct {
	DockerImageRef string
	DockerHost     string
	Kubeconfig     string
}{}

/**
Pulls the latest Semantically Versioned tag of the manifest builder
*/
func pullManifestBuilderImage(ctx context.Context, cli *client.Client) {
	imageRef := fmt.Sprintf("%s:%s", imbImageName, imbTargetVersion)
	out, err := cli.ImagePull(ctx, imageRef, types.ImagePullOptions{})
	if err != nil {
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
	hostConfig := container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:     mount.TypeBind,
				Source:   kubeDir,
				Target:   "/root/.kube",
				ReadOnly: true,
			},
		},
	}
	supplementalDirs := []string{"~/.aws", "~/.minikube"}
	for _, dir := range supplementalDirs {
		sourcePath, err := homedir.Expand(dir)
		if err == nil {
			if _, err := os.Stat(sourcePath); !os.IsNotExist(err) {
				targetPath := fmt.Sprintf("/root/%s", filepath.Base(sourcePath))
				hostConfig.Mounts = append(hostConfig.Mounts,
					mount.Mount{
						Type:     mount.TypeBind,
						Source:   sourcePath,
						Target:   targetPath,
						ReadOnly: true,
					},
					// FIXME: This is some temporary black magic to handle absolute paths in Minikube config
					mount.Mount{
						Type:     mount.TypeBind,
						Source:   sourcePath,
						Target:   sourcePath,
						ReadOnly: true,
					})
			}
		}
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        imageRef,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		OpenStdin:    true,
		StdinOnce:    false,
	}, &hostConfig, nil, "")
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

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNextExit)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case s := <-statusCh:
		if s.StatusCode == 0 {
			copyArtifactsFromContainerToHost(ctx, cli, resp.ID)
		}
	}
	terminal.Restore(fd, oldState)

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
}

func copyArtifactsFromContainerToHost(ctx context.Context, cli *client.Client, containerID string) {
	srcPath := "/app/servo-manifests"
	content, _, err := cli.CopyFromContainer(ctx, containerID, srcPath)
	if err != nil {
		panic(err)
	}
	defer content.Close()

	srcInfo := archive.CopyInfo{
		Path:   srcPath,
		Exists: true,
		IsDir:  true,
	}

	pwd, err := os.Getwd()
	if err != nil {
		pretty.Println("Unable to determine pwd for output path")
		panic(err)
	}
	archive.CopyTo(content, srcInfo, pwd)
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func runDiscovery(args []string) {
	ctx := context.Background()
	clientCfg, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()

	var clusterNames []string
	for _, value := range clientCfg.Contexts {
		clusterNames = append(clusterNames, value.Cluster)
	}
	kubeContext := struct {
		Context    string
		Namespace  string
		Deployment string
	}{}

	var clusterQ = []*survey.Question{
		{
			Name: "Context",
			Prompt: &survey.Select{
				Message: "Select the cluster to be optimized:",
				Options: clusterNames,
			},
			Validate: survey.Required,
		},
	}

	err = survey.Ask(clusterQ, &kubeContext)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("Activating context: %s.\n", kubeContext.Context)

	config, err := clientcmd.BuildConfigFromFlags("", discoverConfig.Kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	namespaces, _ := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	fmt.Printf("There are %d namespaces in the cluster\n", len(namespaces.Items))

	var namespaceNames []string
	for _, value := range namespaces.Items {
		namespaceNames = append(namespaceNames, value.Name)
	}
	var namespaceQ = []*survey.Question{
		{
			Name: "Namespace",
			Prompt: &survey.Select{
				Message: "Select the namespace to be optimized:",
				Options: namespaceNames,
			},
			Validate: survey.Required,
		},
	}
	err = survey.Ask(namespaceQ, &kubeContext)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// Deployments
	deployments, _ := clientset.AppsV1().Deployments(kubeContext.Namespace).List(ctx, metav1.ListOptions{})
	fmt.Printf("There are %d deployments in the namespace\n", len(deployments.Items))

	var deploymentNames []string
	for _, value := range deployments.Items {
		deploymentNames = append(deploymentNames, value.Name)
	}
	var deploymentQ = []*survey.Question{
		{
			Name: "Deployment",
			Prompt: &survey.Select{
				Message: "Select the deployment to be optimized:",
				Options: deploymentNames,
			},
			Validate: survey.Required,
		},
	}
	err = survey.Ask(deploymentQ, &kubeContext)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var clientOpts []client.Opt
	clientOpts = append(clientOpts,
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)

	// Resolve the Docker host using the connection helpers
	// This supports the resolution of ssh:// URL schemes for tunneled execution
	if discoverConfig.DockerHost != "" {
		helper, err := connhelper.GetConnectionHelper(discoverConfig.DockerHost)
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

var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Build Servo assets through Kubernetes discovery",
	Long: `The discover command introspects your Kubernetes and Prometheus
clusters to auto-detect configuration necessary to build a Servo.

Upon completion of discovery, manifests will be generated that can be
used to build a Servo assembly image and deploy it to Kubernetes.`,
	Run: func(cmd *cobra.Command, args []string) {
		if discoverConfig.Kubeconfig == "" {
			discoverConfig.Kubeconfig = pathToDefaultKubeconfig()
		}

		runDiscovery(args)
	},
}

func pathToDefaultKubeconfig() string {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return filepath.Join(home, ".kube", "config")
}

func init() {
	rootCmd.AddCommand(discoverCmd)

	defaultImageRef := fmt.Sprintf("%s:%s", imbImageName, imbTargetVersion)
	discoverCmd.Flags().StringVarP(&discoverConfig.DockerImageRef, "image", "i", defaultImageRef, "Docker image ref to use for discovery")
	discoverCmd.Flags().StringVarP(&discoverConfig.DockerHost, "host", "H", "", "Docket host to connect to (overriding DOCKER_HOST)")
	discoverCmd.Flags().StringVar(&discoverConfig.Kubeconfig, "kubeconfig", "", fmt.Sprintf("Location of the kubeconfig file (default is \"%s\")", pathToDefaultKubeconfig()))
}
