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

package minikube

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"knative.dev/client/pkg/quickstart/install"
)

// NOTE: If you are changing kubernetesVersion and kindVersion, please also
// update the kubectl and kind versions listed here:
// https://github.com/knative-extensions/kn-plugin-quickstart/blob/main/README.md
//
// NOTE: Latest minimum k8 version needed for knative can be found here:
// https://github.com/knative/pkg/blob/main/version/version.go#L36
var kubernetesVersion = "1.32.0"
var clusterName string
var clusterVersionOverride bool
var minikubeVersion = 1.35
var cpus = "3"
var memory = "3072"
var installKnative = true

// SetUp creates a local Minikube cluster and installs all the relevant Knative components
func SetUp(name, kVersion string, installServing, installEventing bool) error {
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
		kubernetesVersion = kVersion
		clusterVersionOverride = true
	}

	if err := createMinikubeCluster(); err != nil {
		return fmt.Errorf("failed to create minikube cluster: %w", err)
	}
	fmt.Print("\n")
	fmt.Println("To finish setting up networking for minikube, run the following command in a separate terminal window:")
	fmt.Println("    minikube tunnel --profile knative")
	fmt.Println("The tunnel command must be running in a terminal window any time when using the knative quickstart environment.")
	fmt.Println("\nPress the Enter key to continue")
	fmt.Scanln()
	if installKnative {
		if installServing {
			if err := install.Serving(""); err != nil {
				return fmt.Errorf("failed to install serving to minikube cluster %s: %w", clusterName, err)
			}
			if err := install.Kourier(); err != nil {
				return fmt.Errorf("failed to install kourier to minikube cluster %s: %w", clusterName, err)
			}
			if err := install.KourierMinikube(); err != nil {
				return fmt.Errorf("failed while configuring kourier for minikube cluster %s: %w", clusterName, err)
			}
		}
		if installEventing {
			if err := install.Eventing(); err != nil {
				return fmt.Errorf("failed to install eventing to minikube cluster %s: %w", clusterName, err)
			}
		}
	}

	finish := time.Since(start).Round(time.Second)
	fmt.Printf("🚀 Knative install took: %s \n", finish)
	fmt.Println("🎉 Now have some fun with Serverless and Event Driven Apps!")

	return nil
}

func createMinikubeCluster() error {
	if err := checkMinikubeVersion(); err != nil {
		return fmt.Errorf("unable to get minikube version: %w", err)
	}
	if err := checkForExistingCluster(); err != nil {
		return fmt.Errorf("failure while handling or checking for existing minikube cluster: %w", err)
	}
	return nil
}

// checkMinikubeVersion validates that the user has the correct version of Minikube installed.
// If not, it prompts the user to download a newer version before continuing.
func checkMinikubeVersion() error {
	versionCheck := exec.Command("minikube", "version", "--short")
	out, err := versionCheck.CombinedOutput()
	if err != nil {
		return fmt.Errorf("unable to check minikube version: %w", err)
	}
	fmt.Printf("Minikube version is: %s\n", string(out))

	userMinikubeVersion, err := parseMinikubeVersion(string(out))
	if err != nil {
		return fmt.Errorf("unable to parse minikube version: %w", err)
	}
	if userMinikubeVersion < minikubeVersion {
		var resp string
		fmt.Printf("WARNING: We recommend at least Minikube v%.2f, while you are using v%.2f\n", minikubeVersion, userMinikubeVersion)
		fmt.Println("You can download a newer version from https://github.com/kubernetes/minikube/releases/")
		fmt.Print("Continue anyway? (not recommended) [y/N]: ")
		fmt.Scanf("%s", &resp)
		if strings.ToLower(resp) != "y" {
			fmt.Println("Installation stopped. Please upgrade minikube and run again")
			os.Exit(0)
		}
	}

	return nil
}

// checkForExistingCluster checks if the user already has a Minikube cluster. If so, it provides
// the option of deleting the existing cluster and recreating it. If not, it proceeds to
// creating a new cluster
func checkForExistingCluster() error {
	getClusters := exec.Command("minikube", "profile", "list")
	out, err := getClusters.CombinedOutput()
	if err != nil {
		// there are no existing minikube profiles, the listing profiles command will error
		// if there were no profiles, we simply want to create a new one and not stop the install
		// so if the error is the "MK_USAGE_NO_PROFILE" error, we ignore it and continue onwards
		if !strings.Contains(string(out), "MK_USAGE_NO_PROFILE") {
			return fmt.Errorf("failed to get existing minikube profiles: %w", err)
		}
	}
	// TODO Add tests for regex
	r := regexp.MustCompile(clusterName)
	matches := r.Match(out)
	if matches {
		var resp string
		fmt.Print("Knative Cluster " + clusterName + " already installed.\nDelete and recreate [y/N]: ")
		fmt.Scanf("%s", &resp)
		if strings.ToLower(resp) != "y" {
			fmt.Println("Installation skipped")
			checkKnativeNamespace := exec.Command("kubectl", "get", "namespaces")
			output, err := checkKnativeNamespace.CombinedOutput()
			namespaces := string(output)
			if err != nil {
				fmt.Println(string(output))
				return fmt.Errorf("unable to get existing kubernetes namespaces for minikube cluster %s: %w", clusterName, err)
			}
			if strings.Contains(namespaces, "knative") {
				fmt.Print("Knative installation already exists.\nDelete and recreate the cluster [y/N]: ")
				fmt.Scanf("%s", &resp)
				if strings.ToLower(resp) != "y" {
					fmt.Println("Skipping installation")
					installKnative = false
					return nil
				} else {
					if err := recreateCluster(); err != nil {
						return fmt.Errorf("failed while recreating minikube cluster: %w", err)
					}
				}
			}
			return nil
		}
		if err := recreateCluster(); err != nil {
			return fmt.Errorf("failed while recreating minikube cluster: %w", err)
		}
		return nil
	}

	if err := createNewCluster(); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

// createNewCluster creates a new Minikube cluster
func createNewCluster() error {
	fmt.Println("☸ Creating Minikube cluster...")

	if !clusterVersionOverride {
		fmt.Println("\nUsing the standard minikube driver for your system")
		fmt.Println("If you wish to use a different driver, please configure minikube using")
		fmt.Print("    minikube config set driver <your-driver>\n\n")

		// If minikube config kubernetes-version exists, use that instead of our default
		if config, ok := getMinikubeConfig("kubernetes-version"); ok {
			kubernetesVersion = config
		}
	}

	// get user configs for memory/cpus if they exist
	if config, ok := getMinikubeConfig("cpus"); ok {
		cpus = config
	}
	if config, ok := getMinikubeConfig("memory"); ok {
		memory = config
	}

	// create cluster and wait until ready
	createCluster := exec.Command("minikube", "start",
		"--kubernetes-version", kubernetesVersion,
		"--cpus", cpus,
		"--memory", memory,
		"--profile", clusterName,
		"--wait", "all",
		"--insecure-registry", "10.0.0.0/24",
		"--addons=registry")
	if err := runCommandWithOutput(createCluster); err != nil {
		return fmt.Errorf("failed to create new minikube cluster %s: %w", clusterName, err)
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

func parseMinikubeVersion(v string) (float64, error) {
	strippedVersion := strings.TrimLeft(strings.TrimRight(v, "\n"), "v")
	dotVersion := strings.Split(strippedVersion, ".")
	floatVersion, err := strconv.ParseFloat(dotVersion[0]+"."+dotVersion[1], 64)
	if err != nil {
		return 0, err
	}

	return floatVersion, nil
}

func getMinikubeConfig(k string) (string, bool) {
	var ok bool
	getConfig := exec.Command("minikube", "config", "get", k)
	v, err := getConfig.Output()
	if err == nil {
		ok = true
	}
	return strings.TrimRight(string(v), "\n"), ok
}

func recreateCluster() error {
	fmt.Println("deleting cluster...")
	deleteCluster := exec.Command("minikube", "delete", "--profile", clusterName)
	if err := deleteCluster.Run(); err != nil {
		return fmt.Errorf("failed to delete minikube cluster %s: %w", clusterName, err)
	}
	if err := createNewCluster(); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}