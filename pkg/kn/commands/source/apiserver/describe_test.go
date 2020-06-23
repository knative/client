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

	"gotest.tools/assert"

	"knative.dev/client/pkg/sources/v1alpha2"
	"knative.dev/client/pkg/util"
)

func TestSimpleDescribe(t *testing.T) {
	apiServerClient := v1alpha2.NewMockKnAPIServerSourceClient(t, "mynamespace")

	apiServerRecorder := apiServerClient.Recorder()
	sampleSource := createAPIServerSource("testsource", "Event", "v1", "testsa", "Reference", map[string]string{"foo": "bar"}, sinkRef)
	sampleSource.Namespace = "mynamespace"
	apiServerRecorder.GetAPIServerSource("testsource", sampleSource, nil)

	out, err := executeAPIServerSourceCommand(apiServerClient, nil, "describe", "testsource")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(out, "testsource", "testsa", "Reference", "testsvc", "Service (serving.knative.dev/v1)", "Resources", "Event", "v1", "Conditions", "foo", "bar", "mynamespace", "default"))
	assert.Assert(t, util.ContainsNone(out, "URI"))

	apiServerRecorder.Validate()
}

func TestDescribeError(t *testing.T) {
	apiServerClient := v1alpha2.NewMockKnAPIServerSourceClient(t, "mynamespace")

	apiServerRecorder := apiServerClient.Recorder()
	apiServerRecorder.GetAPIServerSource("testsource", nil, errors.New("no apiserver source testsource"))

	out, err := executeAPIServerSourceCommand(apiServerClient, nil, "describe", "testsource")
	assert.ErrorContains(t, err, "testsource")
	assert.Assert(t, util.ContainsAll(out, "Usage", "testsource"))

	apiServerRecorder.Validate()
}

func TestDescribeWithSinkURI(t *testing.T) {
	apiServerClient := v1alpha2.NewMockKnAPIServerSourceClient(t, "mynamespace")

	apiServerRecorder := apiServerClient.Recorder()
	sampleSource := createAPIServerSource("testsource", "Event", "v1", "testsa", "Reference", map[string]string{"foo": "bar"}, sinkURI)
	sampleSource.Namespace = "mynamespace"
	apiServerRecorder.GetAPIServerSource("testsource", sampleSource, nil)

	out, err := executeAPIServerSourceCommand(apiServerClient, nil, "describe", "testsource")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(out, "testsource", "testsa", "Reference", "Resources", "Event", "v1", "Conditions", "foo", "bar", "URI", "https", "foo", "mynamespace"))

	apiServerRecorder.Validate()
}
