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

package apiserver

import (
	"errors"
	"testing"

	"gotest.tools/v3/assert"

	v1 "knative.dev/client/pkg/sources/v1"
	"knative.dev/client/pkg/util"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

var (
	sinkURI = duckv1.Destination{
		URI: &apis.URL{
			Scheme: "https",
			Host:   "foo",
		}}
)

func TestSimpleDescribe(t *testing.T) {
	apiServerClient := v1.NewMockKnAPIServerSourceClient(t, "mynamespace")

	apiServerRecorder := apiServerClient.Recorder()
	sampleSource := createAPIServerSource("testsource", "Event", "v1", "testsa", "Reference", map[string]string{"foo": "bar"}, createSinkv1("testsvc", "default"))
	sampleSource.Namespace = "mynamespace"
	apiServerRecorder.GetAPIServerSource("testsource", sampleSource, nil)

	out, err := executeAPIServerSourceCommand(apiServerClient, nil, "describe", "testsource")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(out, "testsource", "testsa", "Reference", "testsvc", "Service (serving.knative.dev/v1)", "Resources", "Event", "v1", "Conditions", "foo", "bar", "mynamespace", "default"))
	assert.Assert(t, util.ContainsNone(out, "URI"))

	apiServerRecorder.Validate()
}

func TestDescribeMachineReadable(t *testing.T) {
	apiServerClient := v1.NewMockKnAPIServerSourceClient(t, "mynamespace")

	apiServerRecorder := apiServerClient.Recorder()
	sampleSource := createAPIServerSource("testsource", "Event", "v1", "testsa", "Reference", map[string]string{"foo": "bar"}, createSinkv1("testsvc", "default"))
	sampleSource.APIVersion = "sources.knative.dev/v1"
	sampleSource.Kind = "ApiServerSource"
	sampleSource.Namespace = "mynamespace"
	apiServerRecorder.GetAPIServerSource("testsource", sampleSource, nil)

	out, err := executeAPIServerSourceCommand(apiServerClient, nil, "describe", "testsource", "-o", "yaml")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(out, "kind: ApiServerSource", "spec:", "status:", "metadata:"))

	apiServerRecorder.Validate()
}

func TestDescribeError(t *testing.T) {
	apiServerClient := v1.NewMockKnAPIServerSourceClient(t, "mynamespace")

	apiServerRecorder := apiServerClient.Recorder()
	apiServerRecorder.GetAPIServerSource("testsource", nil, errors.New("no apiserver source testsource"))

	out, err := executeAPIServerSourceCommand(apiServerClient, nil, "describe", "testsource")
	assert.ErrorContains(t, err, "testsource")
	assert.Assert(t, util.ContainsAll(out, "Usage", "testsource"))

	apiServerRecorder.Validate()
}

func TestDescribeWithSinkURI(t *testing.T) {
	apiServerClient := v1.NewMockKnAPIServerSourceClient(t, "mynamespace")

	apiServerRecorder := apiServerClient.Recorder()
	sampleSource := createAPIServerSource("testsource", "Event", "v1", "testsa", "Reference", map[string]string{"foo": "bar"}, sinkURI)
	sampleSource.Namespace = "mynamespace"
	apiServerRecorder.GetAPIServerSource("testsource", sampleSource, nil)

	out, err := executeAPIServerSourceCommand(apiServerClient, nil, "describe", "testsource")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(out, "testsource", "testsa", "Reference", "Resources", "Event", "v1", "Conditions", "foo", "bar", "URI", "https", "foo", "mynamespace"))

	apiServerRecorder.Validate()
}
