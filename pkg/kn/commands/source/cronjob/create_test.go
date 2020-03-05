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

package cronjob

import (
	"testing"

	"gotest.tools/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	dynamicfake "knative.dev/client/pkg/dynamic/fake"

	clientsourcesv1alpha1 "knative.dev/client/pkg/eventing/legacysources/v1alpha1"
	"knative.dev/client/pkg/util"
)

func TestSimpleCreateCronJobSource(t *testing.T) {
	mysvc := &servingv1.Service{
		TypeMeta:   v1.TypeMeta{Kind: "Service", APIVersion: "serving.knative.dev/v1"},
		ObjectMeta: v1.ObjectMeta{Name: "mysvc", Namespace: "default"},
	}
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("default", mysvc)

	cronjobClient := clientsourcesv1alpha1.NewMockKnCronJobSourceClient(t)

	cronJobRecorder := cronjobClient.Recorder()
	cronJobRecorder.CreateCronJobSource(createCronJobSource("testsource", "* * * * */2", "maxwell", "mysvc", "mysa", "100m", "128Mi", "200m", "256Mi"), nil)

	out, err := executeCronJobSourceCommand(cronjobClient, dynamicClient, "create", "--sink", "svc:mysvc", "--schedule", "* * * * */2", "--data", "maxwell",
		"--service-account", "mysa", "--requests-cpu", "100m", "--requests-memory", "128Mi", "--limits-cpu", "200m", "--limits-memory", "256Mi", "testsource")
	assert.NilError(t, err, "Source should have been created")
	util.ContainsAll(out, "created", "default", "testsource")

	cronJobRecorder.Validate()
}

func TestNoSinkError(t *testing.T) {
	cronjobClient := clientsourcesv1alpha1.NewMockKnCronJobSourceClient(t)

	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("default")

	out, err := executeCronJobSourceCommand(cronjobClient, dynamicClient, "create", "--sink", "svc:mysvc", "--schedule", "* * * * */2", "--data", "maxwell", "testsource")
	assert.Error(t, err, "services.serving.knative.dev \"mysvc\" not found")
	assert.Assert(t, util.ContainsAll(out, "Usage"))
}

func TestNoSinkGivenError(t *testing.T) {
	out, err := executeCronJobSourceCommand(nil, nil, "create", "--schedule", "* * * * */2", "--data", "maxwell", "testsource")
	assert.ErrorContains(t, err, "sink")
	assert.ErrorContains(t, err, "required")
	assert.Assert(t, util.ContainsAll(out, "Usage", "not set", "required"))
}

func TestNoScheduleGivenError(t *testing.T) {
	out, err := executeCronJobSourceCommand(nil, nil, "create", "--sink", "svc:mysvc", "--data", "maxwell", "testsource")
	assert.ErrorContains(t, err, "schedule")
	assert.ErrorContains(t, err, "required")
	assert.Assert(t, util.ContainsAll(out, "Usage", "not set", "required"))
}

func TestNoNameGivenError(t *testing.T) {
	out, err := executeCronJobSourceCommand(nil, nil, "create", "--sink", "svc:mysvc", "--schedule", "* * * * */2")
	assert.ErrorContains(t, err, "name")
	assert.ErrorContains(t, err, "require")
	assert.Assert(t, util.ContainsAll(out, "Usage", "require", "name"))
}
