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

	"gotest.tools/v3/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	dynamicfake "knative.dev/client/pkg/dynamic/fake"
	clientsourcesv1beta2 "knative.dev/client/pkg/sources/v1"

	"knative.dev/client/pkg/util"
)

func TestSimpleCreatePingSource(t *testing.T) {
	mysvc := &servingv1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: "serving.knative.dev/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "mysvc", Namespace: "default"},
	}
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("default", mysvc)

	pingClient := clientsourcesv1beta2.NewMockKnPingSourceClient(t)

	pingRecorder := pingClient.Recorder()
	pingRecorder.CreatePingSource(createPingSource("testsource", "* * * * */2", "maxwell", "", "mysvc", map[string]string{"bla": "blub", "foo": "bar"}), nil)

	out, err := executePingSourceCommand(pingClient, dynamicClient, "create", "--sink", "ksvc:mysvc", "--schedule", "* * * * */2", "--data", "maxwell", "testsource", "--ce-override", "bla=blub", "--ce-override", "foo=bar")
	assert.NilError(t, err, "Source should have been created")
	assert.Assert(t, util.ContainsAll(out, "created", "default", "testsource"))

	pingRecorder.Validate()
}

func TestNoSinkError(t *testing.T) {
	pingClient := clientsourcesv1beta2.NewMockKnPingSourceClient(t)

	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("default")

	out, err := executePingSourceCommand(pingClient, dynamicClient, "create", "--sink", "ksvc:mysvc", "--schedule", "* * * * */2", "--data", "maxwell", "testsource")
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
	out, err := executePingSourceCommand(nil, nil, "create", "--sink", "ksvc:mysvc", "--schedule", "* * * * */2")
	assert.ErrorContains(t, err, "name")
	assert.ErrorContains(t, err, "require")
	assert.Assert(t, util.ContainsAll(out, "Usage", "require", "name"))
}

func TestDataEncoding(t *testing.T) {
	base64Val := "ZGF0YQ=="
	mysvc := &servingv1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: "serving.knative.dev/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "mysvc", Namespace: "default"},
	}
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("default", mysvc)

	pingClient := clientsourcesv1beta2.NewMockKnPingSourceClient(t)

	pingRecorder := pingClient.Recorder()

	pingRecorder.CreatePingSource(createPingSource("testsource", "* * * * */2", "", base64Val, "mysvc", nil), nil)
	out, err := executePingSourceCommand(pingClient, dynamicClient, "create", "--sink", "ksvc:mysvc", "--schedule", "* * * * */2", "--data", base64Val, "testsource")
	assert.NilError(t, err, "Source should have been created")
	assert.Assert(t, util.ContainsAll(out, "created", "default", "testsource"))

	pingRecorder.CreatePingSource(createPingSource("testsource", "* * * * */2", "", base64Val, "mysvc", nil), nil)
	out, err = executePingSourceCommand(pingClient, dynamicClient, "create", "--sink", "ksvc:mysvc", "--schedule", "* * * * */2", "--data", base64Val, "--encoding", "base64", "testsource")
	assert.NilError(t, err, "Source should have been created")
	assert.Assert(t, util.ContainsAll(out, "created", "default", "testsource"))

	pingRecorder.CreatePingSource(createPingSource("testsource", "* * * * */2", base64Val, "", "mysvc", nil), nil)
	out, err = executePingSourceCommand(pingClient, dynamicClient, "create", "--sink", "ksvc:mysvc", "--schedule", "* * * * */2", "--data", base64Val, "--encoding", "text", "testsource")
	assert.NilError(t, err, "Source should have been created")
	assert.Assert(t, util.ContainsAll(out, "created", "default", "testsource"))

	out, err = executePingSourceCommand(pingClient, dynamicClient, "create", "--sink", "ksvc:mysvc", "--schedule", "* * * * */2", "--data", base64Val, "--encoding", "baseMock", "testsource")
	assert.ErrorContains(t, err, "invalid")
	assert.Assert(t, util.ContainsAll(out, "Usage", "text", "base64"))
}
