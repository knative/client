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

	"gotest.tools/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clientv1alpha2 "knative.dev/client/pkg/sources/v1alpha2"
	"knative.dev/client/pkg/util"
)

func TestSimplePingUpdate(t *testing.T) {
	pingSourceClient := clientv1alpha2.NewMockKnPingSourceClient(t)
	pingRecorder := pingSourceClient.Recorder()
	pingRecorder.GetPingSource("testsource", createPingSource("testsource", "* * * * */1", "maxwell", "mysvc", nil), nil)
	pingRecorder.UpdatePingSource(createPingSource("testsource", "* * * * */3", "maxwell", "mysvc", nil), nil)

	out, err := executePingSourceCommand(pingSourceClient, nil, "update", "--schedule", "* * * * */3", "testsource")
	assert.NilError(t, err)
	util.ContainsAll(out, "updated", "default", "testsource")

	pingRecorder.Validate()
}

func TestSimplePingUpdateCEOverrides(t *testing.T) {
	pingSourceClient := clientv1alpha2.NewMockKnPingSourceClient(t)
	pingRecorder := pingSourceClient.Recorder()
	ceOverrideMap := map[string]string{"bla": "blub", "foo": "bar"}
	ceOverrideMapUpdated := map[string]string{"foo": "baz", "new": "ceoverride"}
	pingRecorder.GetPingSource("testsource", createPingSource("testsource", "* * * * */1", "maxwell", "mysvc", ceOverrideMap), nil)
	pingRecorder.UpdatePingSource(createPingSource("testsource", "* * * * */3", "maxwell", "mysvc", ceOverrideMapUpdated), nil)

	out, err := executePingSourceCommand(pingSourceClient, nil, "update", "--schedule", "* * * * */3", "testsource", "--ce-override", "bla-", "--ce-override", "foo=baz", "--ce-override", "new=ceoverride")
	assert.NilError(t, err)
	util.ContainsAll(out, "updated", "default", "testsource")

	pingRecorder.Validate()
}

func TestUpdateError(t *testing.T) {
	pingClient := clientv1alpha2.NewMockKnPingSourceClient(t, "mynamespace")

	pingRecorder := pingClient.Recorder()
	pingRecorder.GetPingSource("testsource", nil, errors.New("no Ping source testsource"))

	out, err := executePingSourceCommand(pingClient, nil, "update", "testsource")
	assert.ErrorContains(t, err, "testsource")
	util.ContainsAll(out, "Usage", "testsource")

	pingRecorder.Validate()
}

func TestPingUpdateDeletionTimestampNotNil(t *testing.T) {
	pingSourceClient := clientv1alpha2.NewMockKnPingSourceClient(t)
	present := createPingSource("test", "", "", "", nil)
	present.DeletionTimestamp = &v1.Time{Time: time.Now()}
	pingRecorder := pingSourceClient.Recorder()
	pingRecorder.GetPingSource("test", present, nil)

	_, err := executePingSourceCommand(pingSourceClient, nil, "update", "test")
	assert.ErrorContains(t, err, present.Name)
	assert.ErrorContains(t, err, "deletion")
	assert.ErrorContains(t, err, "ping")
}
