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

	v1alpha2 "knative.dev/eventing/pkg/apis/sources/v1alpha2"

	clientv1alpha2 "knative.dev/client/pkg/sources/v1alpha2"
	"knative.dev/client/pkg/util"
)

func TestListPingSource(t *testing.T) {
	pingClient := clientv1alpha2.NewMockKnPingSourceClient(t)

	pingRecorder := pingClient.Recorder()
	cJSource := createPingSource("testsource", "* * * * */2", "maxwell", "mysvc")
	cJSourceList := v1alpha2.PingSourceList{}
	cJSourceList.Items = []v1alpha2.PingSource{*cJSource}

	pingRecorder.ListPingSource(&cJSourceList, nil)

	out, err := executePingSourceCommand(pingClient, nil, "list")
	assert.NilError(t, err, "Sources should be listed")
	util.ContainsAll(out, "NAME", "SCHEDULE", "SINK", "AGE", "CONDITIONS", "READY", "REASON")
	util.ContainsAll(out, "testsource", "* * * * */2", "mysvc")

	pingRecorder.Validate()
}

func TestListPingJobSourceEmpty(t *testing.T) {
	pingClient := clientv1alpha2.NewMockKnPingSourceClient(t)

	pingRecorder := pingClient.Recorder()
	cJSourceList := v1alpha2.PingSourceList{}

	pingRecorder.ListPingSource(&cJSourceList, nil)

	out, err := executePingSourceCommand(pingClient, nil, "list")
	assert.NilError(t, err, "Sources should be listed")
	util.ContainsNone(out, "NAME", "SCHEDULE", "SINK", "AGE", "CONDITIONS", "READY", "REASON")
	util.ContainsAll(out, "No", "ping", "source", "found")

	pingRecorder.Validate()
}
