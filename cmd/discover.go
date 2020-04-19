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

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"context"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/pkg/archive"
	"github.com/mitchellh/go-homedir"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const imbImageName string = "opsani/k8s-imb"
const imbTargetVersion string = "latest" // TODO: Should be 1 for semantically versioned containers

// Configuration options bound via Cobra
var discoverConfig = struct {
	DockerImageRef string
	DockerHost     string
	Kubeconfig     string
}{}

func runIntelligentManifestBuilderContainer(ctx context.Context) error {
	di, err := NewDockerInterface()
	if err != nil {
		return err
	}

	err = di.PullImageWithProgressReporting(ctx, discoverConfig.DockerImageRef)
	if err != nil {
		return err
	}

	// FIXME: These paths are expanded locally but over an ssh transport
	// will resolve locally rather than on the remote host
	kubeDir, err := homedir.Expand("~/.kube")
	if err != nil {
		return err
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

	icc := NewInteractiveContainerConfigWithImageRef(discoverConfig.DockerImageRef)
	icc.HostConfig = &hostConfig
	icc.CompletionCallback = copyArtifactsFromContainerToHost
	return di.RunInteractiveContainer(ctx, icc)
}

func copyArtifactsFromContainerToHost(ctx context.Context, di *DockerInterface, cnt container.ContainerCreateCreatedBody, result container.ContainerWaitOKBody) {
	if result.StatusCode != 0 {
		return
	}

	// Check that our paths are workable
	srcPath := "/work/servo-manifests"
	_, err := di.dockerClient.ContainerStatPath(ctx, cnt.ID, srcPath)
	if err != nil {
		fmt.Println("Unable to find artifacts in container, skipping...")
		return
	}

	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Failed copying artifacts from container:", err)
		return
	}

	// Copy them out
	fmt.Println("Copying artifacts...")
	content, _, err := di.dockerClient.CopyFromContainer(ctx, cnt.ID, srcPath)
	if err != nil {
		fmt.Println("Failed copying artifacts from container:", err)
		return
	}
	defer content.Close()

	srcInfo := archive.CopyInfo{
		Path:   srcPath,
		Exists: true,
		IsDir:  true,
	}

	archive.CopyTo(content, srcInfo, pwd)
}

func runDiscovery(args []string) error {
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
		return err
	}
	fmt.Printf("Activating context: %s.\n", kubeContext.Context)

	config, err := clientcmd.BuildConfigFromFlags("", discoverConfig.Kubeconfig)
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
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
		return err
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
		return err
	}
	return nil
}

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull a Docker image",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		di, err := NewDockerInterface()
		if err != nil {
			return err
		}

		return di.PullImageWithProgressReporting(context.Background(), args[0])
	},
}

var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Build Servo assets through Kubernetes discovery",
	Long: `The discover command introspects your Kubernetes and Prometheus
clusters to auto-detect configuration necessary to build a Servo.

Upon completion of discovery, manifests will be generated that can be
used to build a Servo assembly image and deploy it to Kubernetes.`,
	Args:              cobra.NoArgs,
	PersistentPreRunE: InitRequiredToExecute,
	RunE: func(cmd *cobra.Command, args []string) error {
		if discoverConfig.Kubeconfig == "" {
			discoverConfig.Kubeconfig = pathToDefaultKubeconfig()
		}

		return runDiscovery(args)
	},
}

var imbCmd = &cobra.Command{
	Use:   "imb",
	Short: "Run the intelligent manifest builder under Docker",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runIntelligentManifestBuilderContainer(context.Background())
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
	rootCmd.AddCommand(imbCmd)
	rootCmd.AddCommand(pullCmd)

	defaultImageRef := fmt.Sprintf("%s:%s", imbImageName, imbTargetVersion)
	imbCmd.Flags().StringVarP(&discoverConfig.DockerImageRef, "image", "i", defaultImageRef, "Docker image ref to run")
	imbCmd.Flags().StringVarP(&discoverConfig.DockerHost, "host", "H", "", "Docket host to connect to (overriding DOCKER_HOST)")
	discoverCmd.Flags().StringVar(&discoverConfig.Kubeconfig, "kubeconfig", "", fmt.Sprintf("Location of the kubeconfig file (default is \"%s\")", pathToDefaultKubeconfig()))
	discoverCmd.MarkFlagFilename("kubeconfig")
}
