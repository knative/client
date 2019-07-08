// Copyright © 2019 The Knative Authors
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
	"strconv"

	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	servingv1beta1 "github.com/knative/serving/pkg/apis/serving/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

// Give the configuration all the env var values listed in the given map of
// vars.  Does not touch any environment variables not mentioned, but it can add
// new env vars and change the values of existing ones.
func UpdateEnvVars(template *servingv1alpha1.RevisionTemplateSpec, vars map[string]string) error {
	container, err := extractContainer(template)
	if err != nil {
		return err
	}
	container.Env = updateEnvVarsFromMap(container.Env, vars)
	return nil
}

// Update min and max scale annotation if larger than 0
func UpdateConcurrencyConfiguration(template *servingv1alpha1.RevisionTemplateSpec, minScale int, maxScale int, target int, limit int) {
	if minScale != 0 {
		UpdateAnnotation(template, "autoscaling.knative.dev/minScale", strconv.Itoa(minScale))
	}
	if maxScale != 0 {
		UpdateAnnotation(template, "autoscaling.knative.dev/maxScale", strconv.Itoa(maxScale))
	}
	if target != 0 {
		UpdateAnnotation(template, "autoscaling.knative.dev/target", strconv.Itoa(target))
	}

	if limit != 0 {
		template.Spec.ContainerConcurrency = servingv1beta1.RevisionContainerConcurrencyType(limit)
	}
}

// Updater (or add) an annotation to the given service
func UpdateAnnotation(template *servingv1alpha1.RevisionTemplateSpec, annotation string, value string) {
	annoMap := template.Annotations
	if annoMap == nil {
		annoMap = make(map[string]string)
		template.Annotations = annoMap
	}
	annoMap[annotation] = value
}

// Utility function to translate between the API list form of env vars, and the
// more convenient map form.
func EnvToMap(vars []corev1.EnvVar) (map[string]string, error) {
	result := map[string]string{}
	for _, envVar := range vars {
		_, present := result[envVar.Name]
		if present {
			return nil, fmt.Errorf("env var name present more than once: %v", envVar.Name)
		}
		result[envVar.Name] = envVar.Value
	}
	return result, nil
}

// Update a given image
func UpdateImage(template *servingv1alpha1.RevisionTemplateSpec, image string) error {
	container, err := extractContainer(template)
	if err != nil {
		return err
	}
	container.Image = image
	return nil
}

// UpdateContainerPort updates container with a give port
func UpdateContainerPort(template *servingv1alpha1.RevisionTemplateSpec, port int32) error {
	container, err := extractContainer(template)
	if err != nil {
		return err
	}
	container.Ports = []corev1.ContainerPort{{
		ContainerPort: port,
	}}
	return nil
}

func UpdateResources(template *servingv1alpha1.RevisionTemplateSpec, requestsResourceList corev1.ResourceList, limitsResourceList corev1.ResourceList) error {
	container, err := extractContainer(template)
	if err != nil {
		return err
	}
	if container.Resources.Requests == nil {
		container.Resources.Requests = corev1.ResourceList{}
	}

	for k, v := range requestsResourceList {
		container.Resources.Requests[k] = v
	}

	if container.Resources.Limits == nil {
		container.Resources.Limits = corev1.ResourceList{}
	}

	for k, v := range limitsResourceList {
		container.Resources.Limits[k] = v
	}

	return nil
}

// =======================================================================================

func usesOldV1alpha1ContainerField(revision *servingv1alpha1.RevisionTemplateSpec) bool {
	return revision.Spec.DeprecatedContainer != nil
}

func extractContainer(template *servingv1alpha1.RevisionTemplateSpec) (*corev1.Container, error) {
	if usesOldV1alpha1ContainerField(template) {
		return template.Spec.DeprecatedContainer, nil
	}
	containers := template.Spec.Containers
	if len(containers) == 0 {
		return nil, fmt.Errorf("internal: no container set in spec.template.spec.containers")
	}
	if len(containers) > 1 {
		return nil, fmt.Errorf("internal: can't extract container for updating environment"+
			" variables as the configuration contains "+
			"more than one container (i.e. %d containers)", len(containers))
	}
	return &containers[0], nil
}

func updateEnvVarsFromMap(env []corev1.EnvVar, vars map[string]string) []corev1.EnvVar {
	set := make(map[string]bool)
	for i := range env {
		envVar := &env[i]
		value, present := vars[envVar.Name]
		if present {
			envVar.Value = value
			set[envVar.Name] = true
		}
	}
	for name, value := range vars {
		if !set[name] {
			env = append(
				env,
				corev1.EnvVar{
					Name:  name,
					Value: value,
				})
		}
	}
	return env
}
