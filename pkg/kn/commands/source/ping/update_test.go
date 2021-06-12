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

	sourcesv1beta2 "knative.dev/client/pkg/sources/v1beta2"
	"knative.dev/client/pkg/util"
)

func TestSimplePingUpdate(t *testing.T) {
	pingSourceClient := sourcesv1beta2.NewMockKnPingSourceClient(t)
	pingRecorder := pingSourceClient.Recorder()
	pingRecorder.GetPingSource("testsource", createPingSource("testsource", "* * * * */1", "maxwell", "mysvc", nil), nil)
	pingRecorder.UpdatePingSource(createPingSource("testsource", "* * * * */3", "maxwell", "mysvc", nil), nil)

	out, err := executePingSourceCommand(pingSourceClient, nil, "update", "--schedule", "* * * * */3", "testsource")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(out, "updated", "default", "testsource"))

	pingRecorder.Validate()
}

//TestSimplePingUpdateCEOverrides updates ce override, schedule, data and sink
func TestSimplePingUpdateCEOverrides(t *testing.T) {
	pingSourceClient := sourcesv1beta2.NewMockKnPingSourceClient(t)
	pingRecorder := pingSourceClient.Recorder()
	ceOverrideMap := map[string]string{"bla": "blub", "foo": "bar"}
	ceOverrideMapUpdated := map[string]string{"foo": "baz", "new": "ceoverride"}
	pingRecorder.GetPingSource("testsource", createPingSource("testsource", "* * * * */1", "maxwell", "mysvc", ceOverrideMap), nil)
	pingRecorder.UpdatePingSource(createPingSource("testsource", "* * * * */3", "updated-data", "mysvc", ceOverrideMapUpdated), nil)

	out, err := executePingSourceCommand(pingSourceClient, nil, "update", "--schedule", "* * * * */3", "testsource", "--data", "updated-data", "--ce-override", "bla-", "--ce-override", "foo=baz", "--ce-override", "new=ceoverride")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(out, "updated", "default", "testsource"))

	pingRecorder.Validate()
}

func TestUpdateError(t *testing.T) {
	pingClient := sourcesv1beta2.NewMockKnPingSourceClient(t, "mynamespace")

	pingRecorder := pingClient.Recorder()
	pingRecorder.GetPingSource("", nil, errors.New("name of Ping source required"))
	pingRecorder.GetPingSource("testsource", nil, errors.New("no Ping source testsource"))

	out, err := executePingSourceCommand(pingClient, nil, "update", "")
	assert.Error(t, err, "name of Ping source required")
	out, err = executePingSourceCommand(pingClient, nil, "update", "testsource")
	assert.ErrorContains(t, err, "testsource")
	assert.Assert(t, util.ContainsAll(out, "Usage", "testsource"))

	pingRecorder.Validate()
}

func TestPingUpdateDeletionTimestampNotNil(t *testing.T) {
	pingSourceClient := sourcesv1beta2.NewMockKnPingSourceClient(t)
	present := createPingSource("test", "", "", "", nil)
	present.DeletionTimestamp = &metav1.Time{Time: time.Now()}
	pingRecorder := pingSourceClient.Recorder()
	pingRecorder.GetPingSource("test", present, nil)

	_, err := executePingSourceCommand(pingSourceClient, nil, "update", "test")
	assert.ErrorContains(t, err, present.Name)
	assert.ErrorContains(t, err, "deletion")
	assert.ErrorContains(t, err, "ping")
}
