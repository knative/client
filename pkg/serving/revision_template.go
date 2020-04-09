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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/serving/pkg/apis/autoscaling"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

type Scaling struct {
	Min *int
	Max *int
}

func ContainerOfRevisionTemplate(template *servingv1.RevisionTemplateSpec) (*corev1.Container, error) {
	return ContainerOfRevisionSpec(&template.Spec)
}

func ContainerOfRevisionSpec(revisionSpec *servingv1.RevisionSpec) (*corev1.Container, error) {
	if len(revisionSpec.Containers) == 0 {
		return nil, fmt.Errorf("internal: no container set in spec.template.spec.containers")
	}
	return &revisionSpec.Containers[0], nil
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
	c, err := ContainerOfRevisionSpec(revisionSpec)
	if err != nil {
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
