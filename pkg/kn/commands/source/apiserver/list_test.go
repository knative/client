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

	"gotest.tools/assert"

	knsource_v1alpha1 "knative.dev/client/pkg/eventing/legacysources/v1alpha1"
	"knative.dev/client/pkg/util"
	"knative.dev/eventing/pkg/apis/legacysources/v1alpha1"
)

func TestListAPIServerSource(t *testing.T) {
	apiServerClient := knsource_v1alpha1.NewMockKnAPIServerSourceClient(t)

	apiServerRecorder := apiServerClient.Recorder()
	sampleSource := createAPIServerSource("testsource", "Event", "v1", "testsa", "Ref", "testsvc", false)
	sampleSourceList := v1alpha1.ApiServerSourceList{}
	sampleSourceList.Items = []v1alpha1.ApiServerSource{*sampleSource}

	apiServerRecorder.ListAPIServerSource(&sampleSourceList, nil)

	out, err := executeAPIServerSourceCommand(apiServerClient, nil, "list")
	assert.NilError(t, err, "sources should be listed")
	util.ContainsAll(out, "NAME", "RESOURCES", "SINK", "CONDITIONS", "READY", "REASON")
	util.ContainsAll(out, "testsource", "Eventing:v1:false", "mysvc")

	apiServerRecorder.Validate()
}

func TestListAPIServerSourceEmpty(t *testing.T) {
	apiServerClient := knsource_v1alpha1.NewMockKnAPIServerSourceClient(t)

	apiServerRecorder := apiServerClient.Recorder()
	sampleSourceList := v1alpha1.ApiServerSourceList{}

	apiServerRecorder.ListAPIServerSource(&sampleSourceList, nil)

	out, err := executeAPIServerSourceCommand(apiServerClient, nil, "list")
	assert.NilError(t, err, "Sources should be listed")
	util.ContainsNone(out, "NAME", "RESOURCES", "SINK", "CONDITIONS", "READY", "REASON")
	util.ContainsAll(out, "No", "ApiServer", "source", "found")

	apiServerRecorder.Validate()
}
