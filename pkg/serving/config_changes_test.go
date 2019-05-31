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
	"github.com/knative/serving/pkg/apis/serving/v1beta1"
	"reflect"
	"testing"

	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

func TestUpdateEnvVarsNew(t *testing.T) {
	config, container := getV1alpha1ConfigWithOldFields()
	testUpdateEnvVarsNew(t, config, container)
	assertNoV1alpha1(t, config)

	config, container = getV1alpha1Config()
	testUpdateEnvVarsNew(t, config, container)
	assertNoV1alpha1Old(t, config)
}

func testUpdateEnvVarsNew(t *testing.T, config servingv1alpha1.ConfigurationSpec, container *corev1.Container) {
	env := map[string]string{
		"a": "foo",
		"b": "bar",
	}
	err := UpdateEnvVars(&config, env)
	if err != nil {
		t.Fatal(err)
	}
	found, err := EnvToMap(container.Env)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(env, found) {
		t.Fatalf("Env did not match expected %v found %v", env, found)
	}
}

func TestUpdateEnvVarsAppendOld(t *testing.T) {
	config, container := getV1alpha1ConfigWithOldFields()
	testUpdateEnvVarsAppendOld(t, config, container)
	assertNoV1alpha1(t, config)

	config, container = getV1alpha1Config()
	testUpdateEnvVarsAppendOld(t, config, container)
	assertNoV1alpha1Old(t, config)
}

func testUpdateEnvVarsAppendOld(t *testing.T, config servingv1alpha1.ConfigurationSpec, container *corev1.Container) {
	container.Env = []corev1.EnvVar{
		{Name: "a", Value: "foo"},
	}
	env := map[string]string{
		"b": "bar",
	}
	err := UpdateEnvVars(&config, env)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]string{
		"a": "foo",
		"b": "bar",
	}

	found, err := EnvToMap(container.Env)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, found) {
		t.Fatalf("Env did not match expected %v found %v", env, found)
	}
}

func TestUpdateEnvVarsModify(t *testing.T) {
	config, container := getV1alpha1ConfigWithOldFields()
	testUpdateEnvVarsModify(t, config, container)
	assertNoV1alpha1(t, config)

	config, container = getV1alpha1Config()
	testUpdateEnvVarsModify(t, config, container)
	assertNoV1alpha1Old(t, config)
}

func testUpdateEnvVarsModify(t *testing.T, config servingv1alpha1.ConfigurationSpec, container *corev1.Container) {
	container.Env = []corev1.EnvVar{
		corev1.EnvVar{Name: "a", Value: "foo"}}
	env := map[string]string{
		"a": "fancy",
	}
	err := UpdateEnvVars(&config, env)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]string{
		"a": "fancy",
	}

	found, err := EnvToMap(container.Env)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, found) {
		t.Fatalf("Env did not match expected %v found %v", env, found)
	}
}

func TestUpdateEnvVarsBoth(t *testing.T) {
	config, container := getV1alpha1ConfigWithOldFields()
	testUpdateEnvVarsBoth(t, config, container)
	assertNoV1alpha1(t, config)

	config, container = getV1alpha1Config()
	testUpdateEnvVarsBoth(t, config, container)
	assertNoV1alpha1Old(t, config)
}

func testUpdateEnvVarsBoth(t *testing.T, config servingv1alpha1.ConfigurationSpec, container *corev1.Container) {
	container.Env = []corev1.EnvVar{
		corev1.EnvVar{Name: "a", Value: "foo"},
		corev1.EnvVar{Name: "c", Value: "caroline"}}
	env := map[string]string{
		"a": "fancy",
		"b": "boo",
	}
	err := UpdateEnvVars(&config, env)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]string{
		"a": "fancy",
		"b": "boo",
		"c": "caroline",
	}

	found, err := EnvToMap(container.Env)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, found) {
		t.Fatalf("Env did not match expected %v found %v", env, found)
	}
}

// =========================================================================================================

func getV1alpha1ConfigWithOldFields() (servingv1alpha1.ConfigurationSpec, *corev1.Container) {
	container := &corev1.Container{}
	config := servingv1alpha1.ConfigurationSpec{
		DeprecatedRevisionTemplate: &servingv1alpha1.RevisionTemplateSpec{
			Spec: servingv1alpha1.RevisionSpec{
				DeprecatedContainer: container,
			},
		},
	}
	return config, container
}

func getV1alpha1Config() (servingv1alpha1.ConfigurationSpec, *corev1.Container) {
	containers := []corev1.Container{{}}
	config := servingv1alpha1.ConfigurationSpec{
		Template: &servingv1alpha1.RevisionTemplateSpec{
			Spec: servingv1alpha1.RevisionSpec{
				RevisionSpec: v1beta1.RevisionSpec{
					PodSpec: v1beta1.PodSpec{
						Containers: containers,
					},
				},
			},
		},
	}
	return config, &containers[0]
}

func assertNoV1alpha1Old(t *testing.T, spec servingv1alpha1.ConfigurationSpec) {
	if spec.DeprecatedRevisionTemplate != nil {
		t.Error("Assuming only new v1alphav1 fields but fond spec.revisionTemplate")
	}
}

func assertNoV1alpha1(t *testing.T, config servingv1alpha1.ConfigurationSpec) {
	if config.Template != nil {
		t.Error("Assuming only old v1alphav1 fields but fond spec.template")
	}
}
