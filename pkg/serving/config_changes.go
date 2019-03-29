// Copyright ¬© 2019 The Knative Authors
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
	"fmt"

	corev1 "k8s.io/api/core/v1"

	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
)

// Give the configuration all the env var values listed in the given map of
// vars.  Does not touch any environment variables not mentioned, but it can add
// new env vars and change the values of existing ones.
func UpdateEnvVars(config *servingv1alpha1.ConfigurationSpec, vars map[string]string) error {
	set := make(map[string]bool)
	for i, _ := range config.RevisionTemplate.Spec.Container.Env {
		env_var := &config.RevisionTemplate.Spec.Container.Env[i]
		value, present := vars[env_var.Name]
		if present {
			env_var.Value = value
			set[env_var.Name] = true
		}
	}
	for name, value := range vars {
		if !set[name] {
			config.RevisionTemplate.Spec.Container.Env = append(
				config.RevisionTemplate.Spec.Container.Env,
				corev1.EnvVar{
					Name:  name,
					Value: value,
				})
		}
	}
	return nil

}

// Utility function to translate between the API list form of env vars, and the
// more convenient map form.
func EnvToMap(vars []corev1.EnvVar) (map[string]string, error) {
	result := map[string]string{}
	for _, env_var := range vars {
		_, present := result[env_var.Name]
		if present {
			return nil, fmt.Errorf("Env var name present more than once: %v", env_var.Name)
		}
		result[env_var.Name] = env_var.Value
	}
	return result, nil
}

func UpdateImage(config *servingv1alpha1.ConfigurationSpec, image string) error {
	config.RevisionTemplate.Spec.Container.Image = image
	return nil
}
