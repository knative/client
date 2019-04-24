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

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/resource"

	servinglib "github.com/knative/client/pkg/serving"
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

type ConfigurationEditFlags struct {
	Image         string
	Env           []string
	RequestsFlags RequestsFlags
	LimitsFlags   LimitsFlags
}

type RequestsFlags struct {
	CPU    string
	Memory string
}

type LimitsFlags struct {
	CPU    string
	Memory string
}

func (p *ConfigurationEditFlags) AddFlags(command *cobra.Command) {
	command.Flags().StringVar(&p.Image, "image", "", "Image to run.")
	command.Flags().StringArrayVarP(&p.Env, "env", "e", []string{},
		"Environment variable to set. NAME=value; you may provide this flag "+
			"any number of times to set multiple environment variables.")
	command.Flags().StringVar(&p.RequestsFlags.CPU, "requests-cpu", "", "The requested CPU (e.g., 250m).")
	command.Flags().StringVar(&p.RequestsFlags.Memory, "requests-memory", "", "The requested CPU (e.g., 64Mi).")
	command.Flags().StringVar(&p.LimitsFlags.CPU, "limits-cpu", "", "The limits on the requested CPU (e.g., 1000m).")
	command.Flags().StringVar(&p.LimitsFlags.Memory, "limits-memory", "", "The limits on the requested CPU (e.g., 1024Mi).")
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
	limitsResources, err := p.computeLimitsResources(config)
	if err != nil {
		return err
	}
	requestsResources, err := p.computeRequestsResources(config)
	if err != nil {
		return err
	}
	err = servinglib.UpdateResources(config, requestsResources, limitsResources)
	if err != nil {
		return err
	}
	return nil
}

func (p *ConfigurationEditFlags) computeRequestsResources(config *servingv1alpha1.ConfigurationSpec) (corev1.ResourceList, error) {
	resourceList := corev1.ResourceList{}

	if p.RequestsFlags.CPU != "" {
		requestsCPU, err := resource.ParseQuantity(p.RequestsFlags.CPU)
		if err != nil {
			return corev1.ResourceList{}, err
		}

		resourceList[corev1.ResourceCPU] = requestsCPU
	}

	if p.RequestsFlags.Memory != "" {
		requestsMemory, err := resource.ParseQuantity(p.RequestsFlags.Memory)
		if err != nil {
			return corev1.ResourceList{}, err
		}

		resourceList[corev1.ResourceMemory] = requestsMemory
	}

	return resourceList, nil
}

func (p *ConfigurationEditFlags) computeLimitsResources(config *servingv1alpha1.ConfigurationSpec) (corev1.ResourceList, error) {
	resourceList := corev1.ResourceList{}

	if p.LimitsFlags.CPU != "" {
		limitsCPU, err := resource.ParseQuantity(p.LimitsFlags.CPU)
		if err != nil {
			return corev1.ResourceList{}, err
		}

		resourceList[corev1.ResourceCPU] = limitsCPU
	}

	if p.LimitsFlags.Memory != "" {
		limitsMemory, err := resource.ParseQuantity(p.LimitsFlags.Memory)
		if err != nil {
			return corev1.ResourceList{}, err
		}

		resourceList[corev1.ResourceMemory] = limitsMemory
	}

	return resourceList, nil
}
