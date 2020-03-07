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

package ping

import (
	"testing"

	"gotest.tools/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	dynamicfake "knative.dev/client/pkg/dynamic/fake"
	"knative.dev/client/pkg/sources/v1alpha2"

	"knative.dev/client/pkg/util"
)

func TestSimpleCreatePingSource(t *testing.T) {
	mysvc := &servingv1.Service{
		TypeMeta:   v1.TypeMeta{Kind: "Service", APIVersion: "serving.knative.dev/v1"},
		ObjectMeta: v1.ObjectMeta{Name: "mysvc", Namespace: "default"},
	}
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("default", mysvc)

	pingClient := v1alpha2.NewMockKnPingSourceClient(t)

	pingRecorder := pingClient.Recorder()
	pingRecorder.CreatePingSource(createPingSource("testsource", "* * * * */2", "maxwell", "mysvc"), nil)

	out, err := executePingSourceCommand(pingClient, dynamicClient, "create", "--sink", "svc:mysvc", "--schedule", "* * * * */2", "--data", "maxwell", "testsource")
	assert.NilError(t, err, "Source should have been created")
	util.ContainsAll(out, "created", "default", "testsource")

	pingRecorder.Validate()
}

func TestNoSinkError(t *testing.T) {
	pingClient := v1alpha2.NewMockKnPingSourceClient(t)

	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("default")

	out, err := executePingSourceCommand(pingClient, dynamicClient, "create", "--sink", "svc:mysvc", "--schedule", "* * * * */2", "--data", "maxwell", "testsource")
	assert.Error(t, err, "services.serving.knative.dev \"mysvc\" not found")
	assert.Assert(t, util.ContainsAll(out, "Usage"))
}

func TestNoSinkGivenError(t *testing.T) {
	out, err := executePingSourceCommand(nil, nil, "create", "--schedule", "* * * * */2", "--data", "maxwell", "testsource")
	assert.ErrorContains(t, err, "sink")
	assert.ErrorContains(t, err, "required")
	assert.Assert(t, util.ContainsAll(out, "Usage", "not set", "required"))
}

func TestNoNameGivenError(t *testing.T) {
	out, err := executePingSourceCommand(nil, nil, "create", "--sink", "svc:mysvc", "--schedule", "* * * * */2")
	assert.ErrorContains(t, err, "name")
	assert.ErrorContains(t, err, "require")
	assert.Assert(t, util.ContainsAll(out, "Usage", "require", "name"))
}
