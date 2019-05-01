// Copyright © 2018 The Knative Authors
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

package commands

import (
	"fmt"
	"strings"

	servinglib "github.com/knative/client/pkg/serving"
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type ConfigurationEditFlags struct {
	Image                      string
	Env                        []string
	RequestsFlags, LimitsFlags ResourceFlags
}

type ResourceFlags struct {
	CPU    string
	Memory string
}

func (p *ConfigurationEditFlags) AddUpdateFlags(command *cobra.Command) {
	command.Flags().StringVar(&p.Image, "image", "", "Image to run.")
	command.Flags().StringArrayVarP(&p.Env, "env", "e", []string{},
		"Environment variable to set. NAME=value; you may provide this flag "+
			"any number of times to set multiple environment variables.")
	command.Flags().StringVar(&p.RequestsFlags.CPU, "requests-cpu", "", "The requested CPU (e.g., 250m).")
	command.Flags().StringVar(&p.RequestsFlags.Memory, "requests-memory", "", "The requested CPU (e.g., 64Mi).")
	command.Flags().StringVar(&p.LimitsFlags.CPU, "limits-cpu", "", "The limits on the requested CPU (e.g., 1000m).")
	command.Flags().StringVar(&p.LimitsFlags.Memory, "limits-memory", "", "The limits on the requested CPU (e.g., 1024Mi).")
}

func (p *ConfigurationEditFlags) AddCreateFlags(command *cobra.Command) {
	p.AddUpdateFlags(command)
	command.MarkFlagRequired("image")
}

func (p *ConfigurationEditFlags) Apply(config *servingv1alpha1.ConfigurationSpec) error {
	envMap := map[string]string{}
	for _, pairStr := range p.Env {
		pairSlice := strings.SplitN(pairStr, "=", 2)
		if len(pairSlice) <= 1 {
			return fmt.Errorf(
				"--env argument requires a value that contains the '=' character; got %s",
				pairStr)
		}
		envMap[pairSlice[0]] = pairSlice[1]
	}
	err := servinglib.UpdateEnvVars(config, envMap)
	if err != nil {
		return err
	}
	err = servinglib.UpdateImage(config, p.Image)
	if err != nil {
		return err
	}
	limitsResources, err := p.computeResources(p.LimitsFlags)
	if err != nil {
		return err
	}
	requestsResources, err := p.computeResources(p.RequestsFlags)
	if err != nil {
		return err
	}
	err = servinglib.UpdateResources(config, requestsResources, limitsResources)
	if err != nil {
		return err
	}
	return nil
}

func (p *ConfigurationEditFlags) computeResources(resourceFlags ResourceFlags) (corev1.ResourceList, error) {
	resourceList := corev1.ResourceList{}

	if resourceFlags.CPU != "" {
		cpuQuantity, err := resource.ParseQuantity(resourceFlags.CPU)
		if err != nil {
			return corev1.ResourceList{}, err
		}

		resourceList[corev1.ResourceCPU] = cpuQuantity
	}

	if resourceFlags.Memory != "" {
		memoryQuantity, err := resource.ParseQuantity(resourceFlags.Memory)
		if err != nil {
			return corev1.ResourceList{}, err
		}

		resourceList[corev1.ResourceMemory] = memoryQuantity
	}

	return resourceList, nil
}
