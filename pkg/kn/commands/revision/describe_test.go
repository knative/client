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

package revision

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"gotest.tools/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clienttesting "k8s.io/client-go/testing"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	apiserving "knative.dev/serving/pkg/apis/serving"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/yaml"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/util"
)

const (
	imageDigest = "sha256:1234567890123456789012345678901234567890123456789012345678901234"
)

func fakeRevision(args []string, response *servingv1.Revision) (action clienttesting.Action, output string, err error) {
	knParams := &commands.KnParams{}
	cmd, fakeServing, buf := commands.CreateTestKnCommand(NewRevisionCommand(knParams), knParams)
	fakeServing.AddReactor("*", "*",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			action = a
			return true, response, nil
		})
	cmd.SetArgs(args)
	err = cmd.Execute()
	if err != nil {
		return
	}
	output = buf.String()
	return
}

func TestDescribeRevisionWithNoName(t *testing.T) {
	_, _, err := fakeRevision([]string{"revision", "describe"}, &servingv1.Revision{})
	assert.ErrorContains(t, err, "requires", "name", "revision", "single", "argument")
}

func TestDescribeRevisionYaml(t *testing.T) {
	expectedRevision := servingv1.Revision{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Spec: servingv1.RevisionSpec{
			PodSpec: v1.PodSpec{
				Containers: []v1.Container{{
					Name:  "some-container",
					Image: "knative/test:latest",
				}},
			},
		},
		Status: servingv1.RevisionStatus{
			ServiceName: "foo-service",
		},
	}

	action, data, err := fakeRevision([]string{"revision", "describe", "test-rev", "-o", "yaml"}, &expectedRevision)
	if err != nil {
		t.Fatal(err)
	}

	if action == nil {
		t.Fatal("No action")
	} else if !action.Matches("get", "revisions") {
		t.Fatalf("Bad action %v", action)
	}

	jsonData, err := yaml.YAMLToJSON([]byte(data))
	if err != nil {
		t.Fatal(err)
	}

	var returnedRevision servingv1.Revision
	err = json.Unmarshal(jsonData, &returnedRevision)
	if err != nil {
		t.Fatal(err)
	}

	if !equality.Semantic.DeepEqual(expectedRevision, returnedRevision) {
		t.Fatal("mismatched objects")
	}
}

func TestDescribeRevisionBasic(t *testing.T) {
	expectedRevision := createTestRevision("test-rev", 3)

	action, data, err := fakeRevision([]string{"revision", "describe", "test-rev"}, &expectedRevision)
	if err != nil {
		t.Fatal(err)
	}

	if action == nil {
		t.Fatal("No action")
	} else if !action.Matches("get", "revisions") {
		t.Fatalf("Bad action %v", action)
	}

	assert.Assert(t, util.ContainsAll(data, "Image:", "gcr.io/test/image", "++ Ready", "Port:", "8080"))
	assert.Assert(t, util.ContainsAll(data, "EnvFrom:", "cm:test1, cm:test2"))
}

func createTestRevision(revision string, gen int64) servingv1.Revision {
	labels := make(map[string]string)
	labels[apiserving.ConfigurationGenerationLabelKey] = fmt.Sprintf("%d", gen)

	return servingv1.Revision{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Revision",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              revision,
			Namespace:         "default",
			Generation:        1,
			Labels:            labels,
			Annotations:       make(map[string]string),
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: servingv1.RevisionSpec{
			PodSpec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Image: "gcr.io/test/image",
						Env: []v1.EnvVar{
							{Name: "env1", Value: "eval1"},
							{Name: "env2", Value: "eval2"},
						},
						EnvFrom: []v1.EnvFromSource{
							{ConfigMapRef: &v1.ConfigMapEnvSource{LocalObjectReference: v1.LocalObjectReference{Name: "test1"}}},
							{ConfigMapRef: &v1.ConfigMapEnvSource{LocalObjectReference: v1.LocalObjectReference{Name: "test2"}}},
						},
						Ports: []v1.ContainerPort{
							{ContainerPort: 8080},
						},
					},
				},
			},
		},
		Status: servingv1.RevisionStatus{
			ImageDigest: "gcr.io/test/image@" + imageDigest,
			Status: duckv1.Status{
				Conditions: goodConditions(),
			},
		},
	}
}

func goodConditions() duckv1.Conditions {
	ret := make(duckv1.Conditions, 0)
	ret = append(ret, apis.Condition{
		Type:   apis.ConditionReady,
		Status: v1.ConditionTrue,
		LastTransitionTime: apis.VolatileTime{
			Inner: metav1.Time{Time: time.Now()},
		},
	})
	ret = append(ret, apis.Condition{
		Type:   servingv1.ServiceConditionRoutesReady,
		Status: v1.ConditionTrue,
		LastTransitionTime: apis.VolatileTime{
			Inner: metav1.Time{Time: time.Now()},
		},
	})
	return ret
}
