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
	"testing"

	corev1 "k8s.io/api/core/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"gotest.tools/v3/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"knative.dev/serving/pkg/apis/autoscaling"
)

type scalingInfoTest struct {
	min    string
	max    string
	minRes int
	maxRes int
	e      bool
}

func TestScalingInfo(t *testing.T) {
	sentinel := -0xdead
	tests := []scalingInfoTest{
		{"3", "4", 3, 4, false},
		{"", "5", sentinel, 5, false},
		{"4", "", 4, sentinel, false},
		{"", "", sentinel, sentinel, false},
		{"", "funtimes", sentinel, sentinel, true},
	}
	for _, c := range tests {
		m := metav1.ObjectMeta{}
		m.Annotations = map[string]string{}
		if c.min != "" {
			m.Annotations[autoscaling.MinScaleAnnotationKey] = c.min
		}
		if c.max != "" {
			m.Annotations[autoscaling.MaxScaleAnnotationKey] = c.max
		}
		s, err := ScalingInfo(&m)
		if c.e {
			assert.Assert(t, err != nil)
			continue
		} else {
			assert.NilError(t, err)
		}
		if c.minRes != sentinel {
			assert.Assert(t, s.Min != nil)
			assert.Equal(t, c.minRes, *s.Min)
		} else {
			assert.Assert(t, s.Min == nil)
		}
		if c.maxRes != sentinel {
			assert.Assert(t, s.Max != nil)
			assert.Equal(t, c.maxRes, *s.Max)
		} else {
			assert.Assert(t, s.Max == nil)
		}

	}
}

func TestContainerIndexOfRevisionSpec(t *testing.T) {
	tests := []struct {
		name    string
		revSpec *servingv1.RevisionSpec
		want    int
	}{
		{
			"no container",
			&servingv1.RevisionSpec{},
			-1,
		},
		{
			"1 container",
			&servingv1.RevisionSpec{
				PodSpec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "user-container",
							Ports: []corev1.ContainerPort{{ContainerPort: 80}},
						},
					}}},
			0,
		},
		{
			"2 containers",
			&servingv1.RevisionSpec{
				PodSpec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "sidecar-container-1",
						},
						{
							Name:  "user-container",
							Ports: []corev1.ContainerPort{{ContainerPort: 80}},
						},
						{
							Name: "sidecar-container-2",
						},
					}}},
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainerIndexOfRevisionSpec(tt.revSpec); got != tt.want {
				t.Errorf("ContainerIndexOfRevisionSpec() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnnotations(t *testing.T) {
	m := &metav1.ObjectMeta{}
	m.Annotations = map[string]string{UserImageAnnotationKey: "mockImageVal",
		autoscaling.TargetAnnotationKey:            "1",
		autoscaling.TargetUtilizationPercentageKey: "2",
		autoscaling.WindowAnnotationKey:            "mockWindowVal"}
	assert.Equal(t, "mockImageVal", UserImage(m))
	assert.Equal(t, 1, *ConcurrencyTarget(m))
	assert.Equal(t, 2, *ConcurrencyTargetUtilization(m))
	assert.Equal(t, "mockWindowVal", AutoscaleWindow(m))
}

func TestPort(t *testing.T) {
	revisionSpec := &servingv1.RevisionSpec{
		PodSpec:              corev1.PodSpec{},
		ContainerConcurrency: new(int64),
		TimeoutSeconds:       new(int64),
	}
	assert.Equal(t, (*int32)(nil), Port(revisionSpec))
	revisionSpec.PodSpec.Containers = append(revisionSpec.PodSpec.Containers, corev1.Container{})
	assert.Equal(t, (*int32)(nil), Port(revisionSpec))
	port := corev1.ContainerPort{ContainerPort: 42}
	revisionSpec.PodSpec.Containers[0].Ports = []corev1.ContainerPort{port}
	assert.Equal(t, (int32)(42), *Port(revisionSpec))
}
