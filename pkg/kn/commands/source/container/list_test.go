/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package container

import (
	"testing"

	"knative.dev/eventing/pkg/client/clientset/versioned/scheme"

	"gotest.tools/v3/assert"
	v12 "knative.dev/client/pkg/sources/v1"
	"knative.dev/client/pkg/util"
	v1 "knative.dev/eventing/pkg/apis/sources/v1"
)

func TestListContainerSource(t *testing.T) {
	containerClient := v12.NewMockKnContainerSourceClient(t)

	containerRecorder := containerClient.Recorder()
	sampleSource := createContainerSource("testsource", "docker.io/test/newimg", createSinkv1("svc2", "default"))
	sampleSourceList := v1.ContainerSourceList{}
	sampleSourceList.Items = []v1.ContainerSource{*sampleSource}

	containerRecorder.ListContainerSources(&sampleSourceList, nil)

	out, err := executeContainerSourceCommand(containerClient, nil, "list")
	assert.NilError(t, err, "sources should be listed")
	assert.Assert(t, util.ContainsAll(out, "NAME", "IMAGE", "SINK", "AGE", "CONDITIONS", "READY", "REASON"))
	assert.Assert(t, util.ContainsAll(out, "testsource", "docker.io/test/newimg", "ksvc:svc2"))

	containerRecorder.Validate()
}

func TestListContainerSourceEmpty(t *testing.T) {
	containerClient := v12.NewMockKnContainerSourceClient(t)

	containerRecorder := containerClient.Recorder()
	sampleSourceList := v1.ContainerSourceList{}

	containerRecorder.ListContainerSources(&sampleSourceList, nil)

	out, err := executeContainerSourceCommand(containerClient, nil, "list")
	assert.NilError(t, err, "Sources should be listed")
	assert.Assert(t, util.ContainsNone(out, "NAME", "IMAGE", "SINK", "AGE", "CONDITIONS", "READY", "REASON"))
	assert.Assert(t, util.ContainsAll(out, "No", "Container", "source", "found"))

	containerRecorder.Validate()
}

func TestListContainerSourceEmptyWithJsonOutput(t *testing.T) {
	containerClient := v12.NewMockKnContainerSourceClient(t)

	containerRecorder := containerClient.Recorder()
	sampleSourceList := v1.ContainerSourceList{}
	_ = util.UpdateGroupVersionKindWithScheme(&sampleSourceList, v1.SchemeGroupVersion, scheme.Scheme)
	containerRecorder.ListContainerSources(&sampleSourceList, nil)

	out, err := executeContainerSourceCommand(containerClient, nil, "list", "-o", "json")
	assert.NilError(t, err, "Sources should be listed")
	assert.Assert(t, util.ContainsAll(out, "\"apiVersion\": \"sources.knative.dev/v1\"", "\"items\": []", "\"kind\": \"ContainerSourceList\""))

	containerRecorder.Validate()
}
