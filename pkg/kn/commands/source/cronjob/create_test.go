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
	dynamic_fake "knative.dev/client/pkg/dynamic/fake"
	serving_v1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"

	v1alpha12 "knative.dev/client/pkg/eventing/sources/v1alpha1"
	"knative.dev/client/pkg/util"
)

func TestSimpleCreateCronJobSource(t *testing.T) {
	mysvc := &serving_v1alpha1.Service{
		TypeMeta:   v1.TypeMeta{Kind: "Service", APIVersion: "serving.knative.dev/v1alpha1"},
		ObjectMeta: v1.ObjectMeta{Name: "mysvc", Namespace: "default"},
	}
	dynamicClient := dynamic_fake.CreateFakeKnDynamicClient("default", mysvc)

	cronjobClient := v1alpha12.NewMockKnCronJobSourceClient(t)

	cronJobRecorder := cronjobClient.Recorder()
	cronJobRecorder.CreateCronJobSource(createCronJobSource("testsource", "* * * * */2", "maxwell", "mysvc"), nil)

	out, err := executeCronJobSourceCommand(cronjobClient, dynamicClient, "create", "--sink", "svc:mysvc", "--schedule", "* * * * */2", "--data", "maxwell", "testsource")
	assert.NilError(t, err, "Source should have been created")
	util.ContainsAll(out, "created", "default", "testsource")

	cronJobRecorder.Validate()
}

func TestNoSinkError(t *testing.T) {
	cronjobClient := v1alpha12.NewMockKnCronJobSourceClient(t)

	dynamicClient := dynamic_fake.CreateFakeKnDynamicClient("default")

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
