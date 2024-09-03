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
	"errors"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	dynamicfake "knative.dev/client/pkg/dynamic/fake"
	sourcesv1beta2 "knative.dev/client/pkg/sources/v1beta2"
	"knative.dev/client/pkg/util"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func TestSimplePingUpdate(t *testing.T) {
	mysvc1 := &servingv1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: "serving.knative.dev/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "mysvc1", Namespace: "default"},
	}
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("default", mysvc1)
	pingSourceClient := sourcesv1beta2.NewMockKnPingSourceClient(t)
	pingRecorder := pingSourceClient.Recorder()
	pingRecorder.GetPingSource("testsource", createPingSource("testsource", "* * * * */1", "maxwell", "", "mysvc", nil), nil)
	pingRecorder.UpdatePingSource(createPingSource("testsource", "* * * * */3", "maxwell", "", "mysvc1", nil), nil)

	out, err := executePingSourceCommand(pingSourceClient, dynamicClient, "update", "--schedule", "* * * * */3", "testsource", "--sink", "mysvc1")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(out, "updated", "default", "testsource"))

	pingRecorder.GetPingSource("testsource", createPingSource("testsource", "* * * * */1", "maxwell", "", "mysvc", nil), nil)
	pingRecorder.UpdatePingSource(createPingSource("testsource", "* * * * */3", "", "hello", "mysvc", nil), nil)
	out, err = executePingSourceCommand(pingSourceClient, dynamicClient, "update", "--schedule", "* * * * */3", "testsource", "--data", "hello", "--encoding", "base64")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(out, "updated", "default", "testsource"))

	pingRecorder.GetPingSource("testsource", createPingSource("testsource", "* * * * */1", "maxwell", "", "mysvc", nil), nil)
	pingRecorder.UpdatePingSource(createPingSource("testsource", "* * * * */3", "hello", "", "mysvc", nil), nil)
	out, err = executePingSourceCommand(pingSourceClient, dynamicClient, "update", "--schedule", "* * * * */3", "testsource", "--data", "hello", "--encoding", "text")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(out, "updated", "default", "testsource"))

	pingRecorder.GetPingSource("testsource", createPingSource("testsource", "* * * * */1", "maxwell", "", "mysvc", nil), nil)
	pingRecorder.UpdatePingSource(createPingSource("testsource", "* * * * */3", "", "aGVsbG8=", "mysvc", nil), nil)
	out, err = executePingSourceCommand(pingSourceClient, dynamicClient, "update", "--schedule", "* * * * */3", "testsource", "--data", "aGVsbG8=")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(out, "updated", "default", "testsource"))

	pingRecorder.GetPingSource("testsource", createPingSource("testsource", "* * * * */1", "maxwell", "", "mysvc", nil), nil)
	_, err = executePingSourceCommand(pingSourceClient, dynamicClient, "update", "--schedule", "* * * * */3", "testsource", "--data", "aGVsbG8=", "--encoding", "mockencoding")
	assert.ErrorContains(t, err, "invalid value")

	pingRecorder.Validate()
}

// TestSimplePingUpdateCEOverrides updates ce override, schedule, data and sink
func TestSimplePingUpdateCEOverrides(t *testing.T) {
	pingSourceClient := sourcesv1beta2.NewMockKnPingSourceClient(t)
	pingRecorder := pingSourceClient.Recorder()
	ceOverrideMap := map[string]string{"bla": "blub", "foo": "bar"}
	ceOverrideMapUpdated := map[string]string{"foo": "baz", "new": "ceoverride"}
	pingRecorder.GetPingSource("testsource", createPingSource("testsource", "* * * * */1", "maxwell", "", "mysvc", ceOverrideMap), nil)
	pingRecorder.UpdatePingSource(createPingSource("testsource", "* * * * */3", "updated-data", "", "mysvc", ceOverrideMapUpdated), nil)

	out, err := executePingSourceCommand(pingSourceClient, nil, "update", "--schedule", "* * * * */3", "testsource", "--data", "updated-data", "--ce-override", "bla-", "--ce-override", "foo=baz", "--ce-override", "new=ceoverride")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(out, "updated", "default", "testsource"))

	pingRecorder.Validate()
}

func TestUpdateError(t *testing.T) {
	pingClient := sourcesv1beta2.NewMockKnPingSourceClient(t, "mynamespace")

	pingRecorder := pingClient.Recorder()
	pingRecorder.GetPingSource("testsource", nil, errors.New("no Ping source testsource"))

	out, err := executePingSourceCommand(pingClient, nil, "update", "testsource")
	assert.ErrorContains(t, err, "testsource")
	assert.Assert(t, util.ContainsAll(out, "Usage", "testsource"))

	pingRecorder.Validate()
}

func TestPingUpdateDeletionTimestampNotNil(t *testing.T) {
	pingSourceClient := sourcesv1beta2.NewMockKnPingSourceClient(t)
	present := createPingSource("test", "", "", "", "", nil)
	present.DeletionTimestamp = &metav1.Time{Time: time.Now()}
	pingRecorder := pingSourceClient.Recorder()
	pingRecorder.GetPingSource("test", present, nil)

	_, err := executePingSourceCommand(pingSourceClient, nil, "update", "test")
	assert.ErrorContains(t, err, present.Name)
	assert.ErrorContains(t, err, "deletion")
	assert.ErrorContains(t, err, "ping")
}

func TestPingUpdateErrorForNoArgs(t *testing.T) {
	pingClient := sourcesv1beta2.NewMockKnPingSourceClient(t, "mynamespace")
	out, err := executePingSourceCommand(pingClient, nil, "update")
	assert.ErrorContains(t, err, "required")
	assert.Assert(t, util.ContainsAll(out, "Ping", "name", "required"))
}

func TestPingUpdateNoSinkError(t *testing.T) {
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("default")
	pingClient := sourcesv1beta2.NewMockKnPingSourceClient(t)
	pingRecorder := pingClient.Recorder()

	pingRecorder.GetPingSource("testsource", createPingSource("testsource", "* * * * */1", "maxwell", "", "mysvc", nil), nil)

	out, err := executePingSourceCommand(pingClient, dynamicClient, "update", "testsource", "--sink", "ksvc1")
	assert.ErrorContains(t, err, "not found")
	assert.Assert(t, util.ContainsAll(out, "services.serving.knative.dev", "not found", "ksvc1"))
}
