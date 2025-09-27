// Copyright © 2021 The Knative Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kind

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	dclient "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"knative.dev/client/pkg/quickstart/install"
)

// NOTE: If you are changing kubernetesVersion and kindVersion, please also
// update the kubectl and kind versions listed here:
// https://github.com/knative-extensions/kn-plugin-quickstart/blob/main/README.md
//
// NOTE: Latest minimum k8 version needed for knative can be found here:
// https://github.com/knative/pkg/blob/main/version/version.go#L36
var kubernetesVersion = "kindest/node:v1.32.0"
var clusterName string
var kindVersion = 0.26
var containerRegName = "kind-registry"
var containerRegPort = "5001"
var installKnative = true

// SetUp creates a local Kind cluster and installs all the relevant Knative components
func SetUp(name, kVersion string, installServing, installEventing, installKindRegistry bool, installKindExtraMountHostPath string, installKindExtraMountContainerPath string) error {
	start := time.Now()

	// if neither the "install-serving" or "install-eventing" flags are set,
	// then we assume the user wants to install both serving and eventing
	if !installServing && !installEventing {
		installServing = true
		installEventing = true
	}

	// kubectl is required, fail if not found
	if _, err := exec.LookPath("kubectl"); err != nil {
		fmt.Println("ERROR: kubectl is required for quickstart")
		fmt.Println("Download from https://kubectl.docs.kubernetes.io/installation/kubectl/")
		os.Exit(1)
	}

	clusterName = name
	if kVersion != "" {
		if strings.Contains(kVersion, ":") {
			kubernetesVersion = kVersion
		} else {
			kubernetesVersion = "kindest/node:v" + kVersion
		}
	}

	if err := createKindCluster(installKindRegistry, installKindExtraMountHostPath, installKindExtraMountContainerPath); err != nil {
		return fmt.Errorf("failed to create kind cluster: %w", err)
	}
	if installKnative {
		if installServing {
			// Disable tag resolution for localhost registry, since there's no
			// way to redirect Knative Serving to use the kind-registry name.
			// See https://github.com/knative-extensions/kn-plugin-quickstart/issues/467
			registries := ""
			if installKindRegistry {
				registries = fmt.Sprintf("localhost:%s", container_reg_port)
			}
			if err := install.Serving(registries); err != nil {
				return fmt.Errorf("failed to install serving to kind cluster %s: %w", clusterName, err)
			}
			if err := install.Kourier(); err != nil {
				return fmt.Errorf("failed to install kourier to kind cluster %s: %w", clusterName, err)
			}
			if err := install.KourierKind(); err != nil {
				return fmt.Errorf("failed while configuring kourier for kind cluster %s: %w", clusterName, err)
			}
		}
		if installEventing {
			if err := install.Eventing(); err != nil {
				return fmt.Errorf("failed to install eventing to king cluster %s: %w", clusterName, err)
			}
		}
	}

	finish := time.Since(start).Round(time.Second)
	fmt.Printf("🚀 Knative install took: %s \n", finish)
	fmt.Println("🎉 Now have some fun with Serverless and Event Driven Apps!")
	return nil
}

func createKindCluster(registry bool, extraMountHostPath string, extraMountContainerPath string) error {
	dcli, err := checkDocker()
	if err != nil {
		return err
	}
	fmt.Println("✅ Checking dependencies...")
	if err := checkKindVersion(); err != nil {
		return fmt.Errorf("unable to check kind version: %w", err)
	}
	if registry {
		fmt.Println("💽 Installing local registry...")
		if err := pullLocalRegistryImage(dcli); err != nil {
			return fmt.Errorf("%w", err)
		}
		if err := createLocalRegistry(dcli); err != nil {
			return fmt.Errorf("%w", err)
		}
	} else {
		// temporary warning that registry creation is now opt-in
		// remove in v1.12
		fmt.Println("\nA local registry is no longer created by default.")
		fmt.Print("    To create a local registry, use the --registry flag.\n\n")
	}

	if err := checkForExistingCluster(registry, extraMountHostPath, extraMountContainerPath); err != nil {
		return fmt.Errorf("failed while handling or checking for existing kind cluster: %w", err)
	}

	return nil
}

// checkDocker checks that Docker is running on the users local system.
func checkDocker() (*dclient.Client, error) {
	dcli, err := dclient.NewClientWithOpts(dclient.FromEnv, dclient.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	if _, err := dcli.Info(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to get Docker info: %w", err)
	}

	return dcli, nil
}

func pullLocalRegistryImage(dcli *dclient.Client) error {
	ctx := context.Background()
	iorc, err := dcli.ImagePull(ctx, "docker.io/library/registry:2", image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to create local registry container: %w", err)
	}

	scanner := bufio.NewScanner(iorc)
	for scanner.Scan() {
		var jsonData map[string]interface{}
		if err := json.Unmarshal([]byte(scanner.Text()), &jsonData); err != nil {
			break
		}
		fmt.Printf("%s: %s\n", jsonData["status"], jsonData["id"])
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read image pull response: %w", err)
	}
	return nil
}

func createLocalRegistry(dcli *dclient.Client) error {
	if err := deleteContainerRegistry(dcli); err != nil {
		return fmt.Errorf("failed to delete local registry: %w", err)
	}

	resp, err := dcli.ContainerCreate(context.Background(), &container.Config{
		Image: "registry:2",
	}, &container.HostConfig{
		RestartPolicy: container.RestartPolicy{
			Name: "always",
		},
		PortBindings: nat.PortMap{
			"5000/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: container_reg_port,
				},
			},
		},
		NetworkMode: "bridge",
	}, nil, nil, container_reg_name)
	if err != nil {
		return fmt.Errorf("failed to create local registry container: %w", err)
	}

	if err := dcli.ContainerStart(context.Background(), resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start local registry container: %w", err)
	}
	return nil
}

func connectLocalRegistry(dcli *dclient.Client) error {
	err := patchKindNodes()
	if err != nil {
		return fmt.Errorf("failed to patch kind nodes: %w", err)
	}

	err = dcli.NetworkConnect(context.Background(), "kind", container_reg_name, nil)
	if err != nil {
		return fmt.Errorf("failed to connect local registry to kind network: %w", err)
	}

	cm := fmt.Sprintf(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: local-registry-hosting
  namespace: kube-public
data:
  localRegistryHosting.v1: |
    host: "localhost:%s"
    help: "https://kind.sigs.k8s.io/docs/user/local-registry/"`, container_reg_port)
	createLocalRegistryConfigMap := exec.Command("kubectl", "apply", "-f", "-")

	createLocalRegistryConfigMap.Stdin = strings.NewReader(cm)
	// }
	if err := createLocalRegistryConfigMap.Run(); err != nil {
		return fmt.Errorf("failed to create local registry config map: %w", err)
	}
	return nil
}

// checkKindVersion validates that the user has the correct version of Kind installed.
// If not, it prompts the user to download a newer version before continuing.
func checkKindVersion() error {

	versionCheck := exec.Command("kind", "version", "-q")
	out, err := versionCheck.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to get kind version: %w", err)
	}
	fmt.Printf("    Kind version is: %s", string(out))

	userKindVersion, err := parseKindVersion(string(out))
	if err != nil {
		return fmt.Errorf("unable to parse kind version: %w", err)
	}
	if userKindVersion < kindVersion {
		var resp string
		fmt.Printf("WARNING: We recommend at least Kind v%.2f, while you are using v%.2f\n", kindVersion, userKindVersion)
		fmt.Println("You can download a newer version from https://github.com/kubernetes-sigs/kind/releases")
		fmt.Print("Continue anyway? (not recommended) [y/N]: ")
		fmt.Scanf("%s", &resp)
		if resp == "n" || resp == "N" {
			fmt.Println("Installation stopped. Please upgrade kind and run again")
			os.Exit(0)
		}
	}

	return nil
}

// checkForExistingCluster checks if the user already has a Kind cluster. If so, it provides
// the option of deleting the existing cluster and recreating it. If not, it proceeds to
// creating a new cluster
func checkForExistingCluster(registry bool, extraMountHostPath string, extraMountContainerPath string) error {
	getClusters := exec.Command("kind", "get", "clusters", "-q")
	out, err := getClusters.CombinedOutput()
	if err != nil {
		return fmt.Errorf("unable to get kind clusters: %w", err)
	}
	// TODO Add tests for regex
	r := regexp.MustCompile(fmt.Sprintf(`(?m)^%s\n`, clusterName))
	matches := r.Match(out)
	if matches {
		var resp string
		fmt.Print("\nKnative Cluster kind-" + clusterName + " already installed.\nDelete and recreate [y/N]: ")
		fmt.Scanf("%s", &resp)
		if resp == "y" || resp == "Y" {
			if err := recreateCluster(registry, extraMountHostPath, extraMountContainerPath); err != nil {
				return fmt.Errorf("failed while recreating kind cluster %s: %w", clusterName, err)
			}
		} else {
			fmt.Println("\n    Installation skipped")
			checkKnativeNamespace := exec.Command("kubectl", "get", "namespaces")
			output, err := checkKnativeNamespace.CombinedOutput()
			namespaces := string(output)
			if err != nil {
				fmt.Println(string(output))
				return fmt.Errorf("unable to get kubernetes namspaces for kind cluster %s: %w", clusterName, err)
			}
			if strings.Contains(namespaces, "knative") {
				fmt.Print("Knative installation already exists.\nDelete and recreate the cluster [y/N]: ")
				fmt.Scanf("%s", &resp)
				if resp == "y" || resp == "Y" {
					if err := recreateCluster(registry, extraMountHostPath, extraMountContainerPath); err != nil {
						return fmt.Errorf("failed to recreate kind cluster: %w", err)
					}
				} else {
					fmt.Println("Skipping installation")
					installKnative = false
					return nil
				}
			}
			return nil
		}
	} else {
		dcli, err := dclient.NewClientWithOpts(dclient.FromEnv, dclient.WithAPIVersionNegotiation())
		if err != nil {
			return fmt.Errorf("failed to initialize new api client: %w", err)
		}

		if err := createNewCluster(extraMountHostPath, extraMountContainerPath); err != nil {
			return fmt.Errorf("%w", err)
		}
		if registry {
			if err := createLocalRegistry(dcli); err != nil {
				return fmt.Errorf("%w", err)
			}
			if err := connectLocalRegistry(dcli); err != nil {
				return fmt.Errorf("local-registry: %w", err)
			}
		}
	}

	return nil
}

// recreateCluster recreates a Kind cluster
func recreateCluster(registry bool, extraMountHostPath string, extraMountContainerPath string) error {
	fmt.Println("\n    Deleting cluster...")

	dcli, err := dclient.NewClientWithOpts(dclient.FromEnv, dclient.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to initialize new api client: %w", err)
	}

	deleteCluster := exec.Command("kind", "delete", "cluster", "--name", clusterName)
	if err := deleteCluster.Run(); err != nil {
		return fmt.Errorf("failed to delete kind cluster %s: %w", clusterName, err)
	}
	if err := deleteContainerRegistry(dcli); err != nil {
		return fmt.Errorf("failed to delete container registry: %w", err)
	}
	if err := createNewCluster(extraMountHostPath, extraMountContainerPath); err != nil {
		return fmt.Errorf("%w", err)
	}
	if registry {
		if err := createLocalRegistry(dcli); err != nil {
			return fmt.Errorf("%w", err)
		}
		if err := connectLocalRegistry(dcli); err != nil {
			return fmt.Errorf("unable to connect local-registry: %w", err)
		}
	}
	return nil
}

// createNewCluster creates a new Kind cluster
func createNewCluster(extraMountHostPath string, extraMountContainerPath string) error {
	extraMount := ""
	if extraMountHostPath != "" && extraMountContainerPath != "" {
		extraMount = fmt.Sprintf(`
  extraMounts:
  - hostPath: %s
    containerPath: %s`, extraMountHostPath, extraMountContainerPath)
	}

	if extraMount == "" {
		fmt.Println("☸ Creating Kind cluster...")
	} else {
		fmt.Println("☸ Creating Kind cluster with extraMounts...")
	}

	config := fmt.Sprintf(`
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: %s
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry]
    config_path = "/etc/containerd/certs.d/"
nodes:
- role: control-plane
  image: %s %s
  extraPortMappings:
  - containerPort: 31080
    listenAddress: 0.0.0.0
    hostPort: 80`, clusterName, kubernetesVersion, extraMount)

	createCluster := exec.Command("kind", "create", "cluster", "--wait=120s", "--config=-")
	createCluster.Stdin = strings.NewReader(config)
	if err := runCommandWithOutput(createCluster); err != nil {
		return fmt.Errorf("failed to create kind cluster %s: %w", clusterName, err)
	}

	return nil
}

func patchKindNodes() error {
	getNodes := exec.Command("kind", "get", "nodes", "--name", clusterName)
	out, err := getNodes.Output()
	if err != nil {
		return fmt.Errorf("failed to get kind nodes: %w", err)
	}

	nodes := strings.Split(strings.TrimSpace(string(out)), "\n")
	dcli, err := dclient.NewClientWithOpts(dclient.FromEnv, dclient.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}

	for _, node := range nodes {
		fmt.Println("🔗 Patching node: " + node) // DEBUG
		regConfigDir := fmt.Sprintf("/etc/containerd/certs.d/localhost:%s/", container_reg_port)
		execOpts := container.ExecOptions{
			Cmd:    []string{"sh", "-c", fmt.Sprintf(`mkdir -p %s && echo '[host."http://%s:5000"]' > %shosts.toml`, reg_config_dir, container_reg_name, reg_config_dir)},
			Detach: true,
			Tty:    false,
		}

		execIDResp, err := dcli.ContainerExecCreate(context.Background(), node, execOpts)
		if err != nil {
			return fmt.Errorf("failed to create exec instance on node %s: %w", node, err)
		}

		if err := dcli.ContainerExecStart(context.Background(), execIDResp.ID, container.ExecStartOptions{
			Detach: true,
			Tty:    false,
		}); err != nil {
			return fmt.Errorf("failed to start exec instance on node %s: %w", node, err)
		}
	}
	return nil
}

func runCommandWithOutput(c *exec.Cmd) error {
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("piping output: %w", err)
	}
	fmt.Print("\n")
	return nil
}

func parseKindVersion(v string) (float64, error) {
	strippedVersion := strings.TrimLeft(strings.TrimRight(v, "\n"), "v")
	dotVersion := strings.Split(strippedVersion, ".")
	floatVersion, err := strconv.ParseFloat(dotVersion[0]+"."+dotVersion[1], 64)
	if err != nil {
		return 0, err
	}
	return floatVersion, nil
}

func deleteContainerRegistry(dcli *dclient.Client) error {
	if err := dcli.ContainerRemove(context.Background(), container_reg_name, container.RemoveOptions{Force: true}); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), ": no such container") {
			return nil
		}
		return fmt.Errorf("failed remove registry container: %w", err)
	}
	return nil
}