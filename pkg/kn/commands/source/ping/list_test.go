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

	"knative.dev/eventing/pkg/client/clientset/versioned/scheme"

	"gotest.tools/v3/assert"

	sourcesv1beta2 "knative.dev/eventing/pkg/apis/sources/v1beta2"

	clientv1beta2 "knative.dev/client/pkg/sources/v1beta2"
	"knative.dev/client/pkg/util"
)

func TestListPingSource(t *testing.T) {
	pingClient := clientv1beta2.NewMockKnPingSourceClient(t)

	pingRecorder := pingClient.Recorder()
	cJSource := createPingSource("testsource", "* * * * */2", "maxwell", "mysvc", nil)
	cJSourceList := sourcesv1beta2.PingSourceList{}
	cJSourceList.Items = []sourcesv1beta2.PingSource{*cJSource}

	pingRecorder.ListPingSource(&cJSourceList, nil)

	out, err := executePingSourceCommand(pingClient, nil, "list")
	assert.NilError(t, err, "Sources should be listed")
	assert.Assert(t, util.ContainsAll(out, "NAME", "SCHEDULE", "SINK", "AGE", "CONDITIONS", "READY", "REASON"))
	assert.Assert(t, util.ContainsAll(out, "testsource", "* * * * */2", "mysvc"))

	pingRecorder.Validate()
}

func TestListPingJobSourceEmpty(t *testing.T) {
	pingClient := clientv1beta2.NewMockKnPingSourceClient(t)

	pingRecorder := pingClient.Recorder()
	cJSourceList := sourcesv1beta2.PingSourceList{}

	pingRecorder.ListPingSource(&cJSourceList, nil)

	out, err := executePingSourceCommand(pingClient, nil, "list")
	assert.NilError(t, err, "Sources should be listed")
	assert.Assert(t, util.ContainsNone(out, "NAME", "SCHEDULE", "SINK", "AGE", "CONDITIONS", "READY", "REASON"))
	assert.Assert(t, util.ContainsAll(out, "No", "Ping", "source", "found"))

	pingRecorder.Validate()
}

func TestListPingJobSourceEmptyWithJsonOutput(t *testing.T) {
	pingClient := clientv1beta2.NewMockKnPingSourceClient(t)

	pingRecorder := pingClient.Recorder()
	cJSourceList := sourcesv1beta2.PingSourceList{}
	_ = util.UpdateGroupVersionKindWithScheme(&cJSourceList, sourcesv1beta2.SchemeGroupVersion, scheme.Scheme)
	pingRecorder.ListPingSource(&cJSourceList, nil)

	out, err := executePingSourceCommand(pingClient, nil, "list", "-o", "json")
	assert.NilError(t, err, "Sources should be listed")
	assert.Assert(t, util.ContainsAll(out, "\"apiVersion\": \"sources.knative.dev/v1beta2\"", "\"items\": []", "\"kind\": \"PingSourceList\""))

	pingRecorder.Validate()
}
