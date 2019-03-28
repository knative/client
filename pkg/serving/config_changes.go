// Copyright ¬© 2018 The Knative Authors
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

package serving

import (
	corev1 "k8s.io/api/core/v1"

	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
)

type ConfigChange func(*servingv1alpha1.ConfigurationSpec) error

func EnvVarUpdate(vars map[string]string) ConfigChange {
	return func(config *servingv1alpha1.ConfigurationSpec) error {
		set := make(map[string]bool)
		for _, env_var := range config.RevisionTemplate.Spec.Container.Env {
			value, present := vars[env_var.Name]
			if present {
				env_var.Value = value
				set[env_var.Name] = true
			}
		}
		for name, value := range vars {
			if !set[name] {
				config.RevisionTemplate.Spec.Container.Env = append(config.RevisionTemplate.Spec.Container.Env, corev1.EnvVar{
					Name:  name,
					Value: value,
				})
			}
		}
		return nil
	}
}

func ImageUpdate(image string) ConfigChange {
	return func(config *servingv1alpha1.ConfigurationSpec) error {
		config.RevisionTemplate.Spec.Container.Image = image
		return nil
	}
}
