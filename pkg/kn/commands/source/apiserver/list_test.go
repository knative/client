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
	"testing"

	"knative.dev/eventing/pkg/client/clientset/versioned/scheme"

	"gotest.tools/v3/assert"

	v1alpha2 "knative.dev/eventing/pkg/apis/sources/v1alpha2"

	v1alpha22 "knative.dev/client/pkg/sources/v1alpha2"
	"knative.dev/client/pkg/util"
)

func TestListAPIServerSource(t *testing.T) {
	apiServerClient := v1alpha22.NewMockKnAPIServerSourceClient(t)

	apiServerRecorder := apiServerClient.Recorder()
	sampleSource := createAPIServerSource("testsource", "Event", "v1", "testsa", "Reference", nil, createSinkv1("testsvc", "default"))
	sampleSourceList := v1alpha2.ApiServerSourceList{}
	sampleSourceList.Items = []v1alpha2.ApiServerSource{*sampleSource}

	apiServerRecorder.ListAPIServerSource(&sampleSourceList, nil)

	out, err := executeAPIServerSourceCommand(apiServerClient, nil, "list")
	assert.NilError(t, err, "sources should be listed")
	assert.Assert(t, util.ContainsAll(out, "NAME", "RESOURCES", "SINK", "AGE", "CONDITIONS", "READY", "REASON"))
	assert.Assert(t, util.ContainsAll(out, "testsource", "Event:v1", "ksvc:testsvc"))

	apiServerRecorder.Validate()
}

func TestListAPIServerSourceEmpty(t *testing.T) {
	apiServerClient := v1alpha22.NewMockKnAPIServerSourceClient(t)

	apiServerRecorder := apiServerClient.Recorder()
	sampleSourceList := v1alpha2.ApiServerSourceList{}

	apiServerRecorder.ListAPIServerSource(&sampleSourceList, nil)

	out, err := executeAPIServerSourceCommand(apiServerClient, nil, "list")
	assert.NilError(t, err, "Sources should be listed")
	assert.Assert(t, util.ContainsNone(out, "NAME", "RESOURCES", "SINK", "AGE", "CONDITIONS", "READY", "REASON"))
	assert.Assert(t, util.ContainsAll(out, "No", "ApiServer", "source", "found"))

	apiServerRecorder.Validate()
}

func TestListAPIServerSourceEmptyWithJsonOutput(t *testing.T) {
	apiServerClient := v1alpha22.NewMockKnAPIServerSourceClient(t)

	apiServerRecorder := apiServerClient.Recorder()
	sampleSourceList := v1alpha2.ApiServerSourceList{}
	_ = util.UpdateGroupVersionKindWithScheme(&sampleSourceList, v1alpha2.SchemeGroupVersion, scheme.Scheme)
	apiServerRecorder.ListAPIServerSource(&sampleSourceList, nil)

	out, err := executeAPIServerSourceCommand(apiServerClient, nil, "list", "-o", "json")
	assert.NilError(t, err, "Sources should be listed")
	assert.Assert(t, util.ContainsAll(out, "\"apiVersion\": \"sources.knative.dev/v1alpha2\"", "\"items\": []", "\"kind\": \"ApiServerSourceList\""))

	apiServerRecorder.Validate()
}
