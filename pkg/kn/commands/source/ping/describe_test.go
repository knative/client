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

	"gotest.tools/v3/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	sourcesv1beta "knative.dev/eventing/pkg/apis/sources/v1beta2"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	clientv1beta2 "knative.dev/client/pkg/sources/v1beta2"
	"knative.dev/client/pkg/util"
	"knative.dev/pkg/apis"
)

func TestDescribeRef(t *testing.T) {
	pingClient := clientv1beta2.NewMockKnPingSourceClient(t, "mynamespace")

	pingRecorder := pingClient.Recorder()
	pingRecorder.GetPingSource("testping",
		createPingSource("testping", "*/2 * * * *", "test", "testsvc", map[string]string{"foo": "bar"}), nil)

	out, err := executePingSourceCommand(pingClient, nil, "describe", "testping")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(out, "*/2 * * * *", "test", "testsvc", "Service", "Overrides", "foo", "bar", "Conditions"))

	pingRecorder.Validate()
}

func TestDescribeURI(t *testing.T) {
	pingClient := clientv1beta2.NewMockKnPingSourceClient(t, "mynamespace")

	pingRecorder := pingClient.Recorder()
	pingRecorder.GetPingSource("testsource-uri", getPingSourceSinkURI(), nil)

	out, err := executePingSourceCommand(pingClient, nil, "describe", "testsource-uri")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(out, "mynamespace", "1 2 3 4 5", "honeymoon", "URI", "https", "foo", "testsource-uri"))

	pingRecorder.Validate()
}

func TestDescribeMachineReadable(t *testing.T) {
	pingClient := clientv1beta2.NewMockKnPingSourceClient(t, "mynamespace")

	pingRecorder := pingClient.Recorder()
	pingRecorder.GetPingSource("testsource-uri", getPingSourceSinkURI(), nil)

	out, err := executePingSourceCommand(pingClient, nil, "describe", "testsource-uri", "-o", "yaml")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(out, "kind: PingSource", "spec:", "status:", "metadata:"))
	pingRecorder.Validate()
}

func TestDescribeError(t *testing.T) {
	pingClient := clientv1beta2.NewMockKnPingSourceClient(t, "mynamespace")

	pingRecorder := pingClient.Recorder()
	pingRecorder.GetPingSource("testsource", nil, errors.New("no Ping source testsource"))

	out, err := executePingSourceCommand(pingClient, nil, "describe", "testsource")
	assert.ErrorContains(t, err, "testsource")
	assert.Assert(t, util.ContainsAll(out, "Usage", "testsource"))

	pingRecorder.Validate()

}

func TestPingDescribeErrorForNoArgs(t *testing.T) {
	pingClient := clientv1beta2.NewMockKnPingSourceClient(t, "mynamespace")
	out, err := executePingSourceCommand(pingClient, nil, "describe")
	assert.ErrorContains(t, err, "single argument")
	assert.Assert(t, util.ContainsAll(out, "requires", "single argument"))
}

func getPingSourceSinkURI() *sourcesv1beta.PingSource {
	return &sourcesv1beta.PingSource{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PingSource",
			APIVersion: "sources.knative.dev/v1beta2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testsource-uri",
			Namespace: "mynamespace",
		},
		Spec: sourcesv1beta.PingSourceSpec{
			Schedule: "1 2 3 4 5",
			Data:     "honeymoon",
			SourceSpec: duckv1.SourceSpec{
				Sink: duckv1.Destination{
					URI: &apis.URL{
						Scheme: "https",
						Host:   "foo",
					},
				},
			},
		},
		Status: sourcesv1beta.PingSourceStatus{},
	}
}
