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
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/knative/serving/pkg/apis/autoscaling"
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	servingv1beta1 "github.com/knative/serving/pkg/apis/serving/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

// UpdateEnvVars gives the configuration all the env var values listed in the given map of
// vars.  Does not touch any environment variables not mentioned, but it can add
// new env vars and change the values of existing ones.
func UpdateEnvVars(template *servingv1alpha1.RevisionTemplateSpec, toUpdate map[string]string, toRemove []string) error {
	container, err := ContainerOfRevisionTemplate(template)
	if err != nil {
		return err
	}
	envVars := updateEnvVarsFromMap(container.Env, toUpdate)
	container.Env = removeEnvVars(envVars, toRemove)
	return nil
}

// UpdateMinScale updates min scale annotation
func UpdateMinScale(template *servingv1alpha1.RevisionTemplateSpec, min int) error {
	return UpdateAnnotation(template, autoscaling.MinScaleAnnotationKey, strconv.Itoa(min))
}

// UpdatMaxScale updates max scale annotation
func UpdateMaxScale(template *servingv1alpha1.RevisionTemplateSpec, max int) error {
	return UpdateAnnotation(template, autoscaling.MaxScaleAnnotationKey, strconv.Itoa(max))
}

// UpdateConcurrencyTarget updates container concurrency annotation
func UpdateConcurrencyTarget(template *servingv1alpha1.RevisionTemplateSpec, target int) error {
	// TODO(toVersus): Remove the following validation once serving library is updated to v0.8.0
	// and just rely on ValidateAnnotations method.
	if target < autoscaling.TargetMin {
		return fmt.Errorf("Invalid 'concurrency-target' value: must be an integer greater than 0: %s",
			autoscaling.TargetAnnotationKey)
	}

	return UpdateAnnotation(template, autoscaling.TargetAnnotationKey, strconv.Itoa(target))
}

// UpdateConcurrencyLimit updates container concurrency limit
func UpdateConcurrencyLimit(template *servingv1alpha1.RevisionTemplateSpec, limit int) error {
	cc := servingv1beta1.RevisionContainerConcurrencyType(limit)
	// Validate input limit
	ctx := context.Background()
	if err := cc.Validate(ctx).ViaField("spec.containerConcurrency"); err != nil {
		return fmt.Errorf("Invalid 'concurrency-limit' value: %s", err)
	}
	template.Spec.ContainerConcurrency = cc
	return nil
}

// UpdateAnnotation updates (or adds) an annotation to the given service
func UpdateAnnotation(template *servingv1alpha1.RevisionTemplateSpec, annotation string, value string) error {
	annoMap := template.Annotations
	if annoMap == nil {
		annoMap = make(map[string]string)
		template.Annotations = annoMap
	}

	// Validate autoscaling annotations and returns error if invalid input provided
	// without changing the existing spec
	in := make(map[string]string)
	in[annotation] = value
	if err := autoscaling.ValidateAnnotations(in); err != nil {
		return err
	}

	annoMap[annotation] = value
	return nil
}

var ApiTooOldError = errors.New("the service is using too old of an API format for the operation")

// UpdateName updates the revision name.
func UpdateName(template *servingv1alpha1.RevisionTemplateSpec, name string) error {
	if template.Spec.DeprecatedContainer != nil {
		return ApiTooOldError
	}
	template.Name = name
	return nil
}

// EnvToMap is an utility function to translate between the API list form of env vars, and the
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

// UpdateImage a given image
func UpdateImage(template *servingv1alpha1.RevisionTemplateSpec, image string) error {
	container, err := ContainerOfRevisionTemplate(template)
	if err != nil {
		return err
	}
	container.Image = image
	return nil
}

// UpdateContainerPort updates container with a give port
func UpdateContainerPort(template *servingv1alpha1.RevisionTemplateSpec, port int32) error {
	container, err := ContainerOfRevisionTemplate(template)
	if err != nil {
		return err
	}
	container.Ports = []corev1.ContainerPort{{
		ContainerPort: port,
	}}
	return nil
}

// UpdateResources updates resources as requested
func UpdateResources(template *servingv1alpha1.RevisionTemplateSpec, requestsResourceList corev1.ResourceList, limitsResourceList corev1.ResourceList) error {
	container, err := ContainerOfRevisionTemplate(template)
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

// UpdateLabels updates the labels identically on a service and template.
// Does not overwrite the entire Labels field, only makes the requested updates
func UpdateLabels(service *servingv1alpha1.Service, template *servingv1alpha1.RevisionTemplateSpec, toUpdate map[string]string, toRemove []string) error {
	if service.ObjectMeta.Labels == nil {
		service.ObjectMeta.Labels = make(map[string]string)
	}
	if template.ObjectMeta.Labels == nil {
		template.ObjectMeta.Labels = make(map[string]string)
	}
	for key, value := range toUpdate {
		service.ObjectMeta.Labels[key] = value
		template.ObjectMeta.Labels[key] = value
	}
	for _, key := range toRemove {
		delete(service.ObjectMeta.Labels, key)
		delete(template.ObjectMeta.Labels, key)
	}
	return nil
}

// =======================================================================================

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

func removeEnvVars(env []corev1.EnvVar, toRemove []string) []corev1.EnvVar {
	for _, name := range toRemove {
		for i, envVar := range env {
			if envVar.Name == name {
				env = append(env[:i], env[i+1:]...)
				break
			}
		}
	}
	return env
}
