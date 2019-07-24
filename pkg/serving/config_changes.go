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
func UpdateEnvVars(template *servingv1alpha1.RevisionTemplateSpec, vars map[string]string) error {
	container, err := ContainerOfRevisionTemplate(template)
	if err != nil {
		return err
	}
	container.Env = updateEnvVarsFromMap(container.Env, vars)
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
		return fmt.Errorf("Invalid %s annotation value: must be an intger greater than 0", autoscaling.TargetAnnotationKey)
	}

	return UpdateAnnotation(template, autoscaling.TargetAnnotationKey, strconv.Itoa(target))
}

// UpdateConcurrencyLimit updates container concurrency limit
func UpdateConcurrencyLimit(template *servingv1alpha1.RevisionTemplateSpec, limit int) error {
	cc := servingv1beta1.RevisionContainerConcurrencyType(limit)
	// Validate input limit
	ctx := context.Background()
	if err := cc.Validate(ctx).ViaField("spec.containerConcurrency"); err != nil {
		return fmt.Errorf("Invalid containerConcurrency revision spec: %s", err)
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

	// Validate autoscaling annotations and returns the same value as before if input value is invalid
	in := make(map[string]string)
	in[annotation] = value
	if err := autoscaling.ValidateAnnotations(in); err != nil {
		return err
	}

	annoMap[annotation] = value
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
