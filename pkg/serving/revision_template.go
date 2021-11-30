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
	"strconv"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/serving/pkg/apis/autoscaling"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

type Scaling struct {
	Min *int
	Max *int
}

// ContainerOfRevisionSpec returns the 'main' container of a revision specification and
// use GetServingContainerIndex to identify the container.
// Nil is returned if no such container could be found
func ContainerOfRevisionSpec(revisionSpec *servingv1.RevisionSpec) *corev1.Container {
	idx := ContainerIndexOfRevisionSpec(revisionSpec)
	if idx == -1 {
		return nil
	}
	return &revisionSpec.Containers[0]
}

// ContainerIndexOfRevisionSpec returns the index of the "main" container if
// multiple containers are present. The main container is either the single
// container when there is only ony container in the list or the first container
// which has a ports declaration (validation guarantees that there is only one
// such container)
// If no container could be found (list is empty or no container has a port declaration)
// then -1 is returned
// This method's logic is taken from RevisionSpec.GetContainer()
func ContainerIndexOfRevisionSpec(revisionSpec *servingv1.RevisionSpec) int {
	switch {
	case len(revisionSpec.Containers) == 1:
		return 0
	case len(revisionSpec.Containers) > 1:
		for i := range revisionSpec.Containers {
			if len(revisionSpec.Containers[i].Ports) != 0 {
				return i
			}
		}
	}
	return -1
}

// ContainerStatus returns the status of the main container or nil of no
// such status could be found
func ContainerStatus(r *servingv1.Revision) *servingv1.ContainerStatus {
	idx := ContainerIndexOfRevisionSpec(&r.Spec)
	if idx < 0 || idx >= len(r.Status.ContainerStatuses) {
		return nil
	}
	return &r.Status.ContainerStatuses[idx]
}

func ScalingInfo(m *metav1.ObjectMeta) (*Scaling, error) {
	ret := &Scaling{}
	var err error
	ret.Min, err = annotationAsInt(m, autoscaling.MinScaleAnnotationKey)
	if err != nil {
		return nil, err
	}
	ret.Max, err = annotationAsInt(m, autoscaling.MaxScaleAnnotationKey)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func UserImage(m *metav1.ObjectMeta) string {
	return m.Annotations[UserImageAnnotationKey]
}

func ConcurrencyTarget(m *metav1.ObjectMeta) *int {
	ret, _ := annotationAsInt(m, autoscaling.TargetAnnotationKey)
	return ret
}

func ConcurrencyTargetUtilization(m *metav1.ObjectMeta) *int {
	ret, _ := annotationAsInt(m, autoscaling.TargetUtilizationPercentageKey)
	return ret
}

func AutoscaleWindow(m *metav1.ObjectMeta) string {
	return m.Annotations[autoscaling.WindowAnnotationKey]
}

func Port(revisionSpec *servingv1.RevisionSpec) *int32 {
	c := ContainerOfRevisionSpec(revisionSpec)
	if c == nil {
		return nil
	}
	if len(c.Ports) > 0 {
		p := c.Ports[0].ContainerPort
		return &p
	}
	return nil
}

// =======================================================================================

func annotationAsInt(m *metav1.ObjectMeta, annotationKey string) (*int, error) {
	annos := m.Annotations
	if val, ok := annos[annotationKey]; ok {
		valInt, err := strconv.Atoi(val)
		if err != nil {
			return nil, err
		}
		return &valInt, nil
	}
	return nil, nil
}
