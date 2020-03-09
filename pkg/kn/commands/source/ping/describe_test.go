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

	"gotest.tools/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1alpha2 "knative.dev/eventing/pkg/apis/sources/v1alpha2"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	clientv1alpha2 "knative.dev/client/pkg/sources/v1alpha2"
	"knative.dev/client/pkg/util"
)

func TestSimpleDescribe(t *testing.T) {
	pingClient := clientv1alpha2.NewMockKnPingSourceClient(t, "mynamespace")

	pingRecorder := pingClient.Recorder()
	pingRecorder.GetPingSource("testsource", getPingSource(), nil)

	out, err := executePingSourceCommand(pingClient, nil, "describe", "testsource")
	assert.NilError(t, err)
	util.ContainsAll(out, "1 2 3 4 5", "honeymoon", "myservicenamespace", "mysvc", "Service", "testsource")
	util.ContainsAll(out, "myaccount", "100m", "128Mi", "200m", "256Mi")

	pingRecorder.Validate()

}

func TestDescribeError(t *testing.T) {
	pingClient := clientv1alpha2.NewMockKnPingSourceClient(t, "mynamespace")

	pingRecorder := pingClient.Recorder()
	pingRecorder.GetPingSource("testsource", nil, errors.New("no Ping source testsource"))

	out, err := executePingSourceCommand(pingClient, nil, "describe", "testsource")
	assert.ErrorContains(t, err, "testsource")
	util.ContainsAll(out, "Usage", "testsource")

	pingRecorder.Validate()

}

func getPingSource() *v1alpha2.PingSource {
	return &v1alpha2.PingSource{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: "testsource",
		},
		Spec: v1alpha2.PingSourceSpec{
			Schedule: "1 2 3 4 5",
			JsonData: "honeymoon",
			SourceSpec: duckv1.SourceSpec{
				Sink: duckv1.Destination{
					Ref: &duckv1.KReference{
						Kind:      "Service",
						Namespace: "myservicenamespace",
						Name:      "mysvc",
					},
				},
			},
		},
		Status: v1alpha2.PingSourceStatus{},
	}
}
