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
	"reflect"
	"testing"

	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

func TestUpdateEnvVarsNew(t *testing.T) {
	config := getEmptyConfigurationSpec()
	env := map[string]string{
		"a": "foo",
		"b": "bar",
	}
	err := UpdateEnvVars(&config, env)
	if err != nil {
		t.Fatal(err)
	}
	found, err := EnvToMap(config.DeprecatedRevisionTemplate.Spec.DeprecatedContainer.Env)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(env, found) {
		t.Fatalf("Env did not match expected %v found %v", env, found)
	}
}

func TestUpdateEnvVarsAppend(t *testing.T) {
	config := getEmptyConfigurationSpec()
	config.DeprecatedRevisionTemplate.Spec.DeprecatedContainer.Env = []corev1.EnvVar{
		corev1.EnvVar{Name: "a", Value: "foo"}}
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

	found, err := EnvToMap(config.DeprecatedRevisionTemplate.Spec.DeprecatedContainer.Env)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, found) {
		t.Fatalf("Env did not match expected %v found %v", env, found)
	}
}

func TestUpdateEnvVarsModify(t *testing.T) {
	config := getEmptyConfigurationSpec()
	config.DeprecatedRevisionTemplate.Spec.DeprecatedContainer.Env = []corev1.EnvVar{
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

	found, err := EnvToMap(config.DeprecatedRevisionTemplate.Spec.DeprecatedContainer.Env)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, found) {
		t.Fatalf("Env did not match expected %v found %v", env, found)
	}
}

func TestUpdateEnvVarsBoth(t *testing.T) {
	config := getEmptyConfigurationSpec()
	config.DeprecatedRevisionTemplate.Spec.DeprecatedContainer.Env = []corev1.EnvVar{
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

	found, err := EnvToMap(config.DeprecatedRevisionTemplate.Spec.DeprecatedContainer.Env)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, found) {
		t.Fatalf("Env did not match expected %v found %v", env, found)
	}
}

func getEmptyConfigurationSpec() servingv1alpha1.ConfigurationSpec {
	return servingv1alpha1.ConfigurationSpec{
		DeprecatedRevisionTemplate: &servingv1alpha1.RevisionTemplateSpec{
			Spec: servingv1alpha1.RevisionSpec{
				DeprecatedContainer: &corev1.Container{},
			},
		},
	}
}
