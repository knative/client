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
	"errors"
	"testing"

	"gotest.tools/v3/assert"
	"knative.dev/client/pkg/sources/v1alpha2"
	"knative.dev/client/pkg/util"
)

func TestDescribeError(t *testing.T) {
	containerClient := v1alpha2.NewMockKnContainerSourceClient(t, "mynamespace")

	containerRecorder := containerClient.Recorder()
	containerRecorder.GetContainerSource("testsource", nil, errors.New("no container source testsource"))

	out, err := executeContainerSourceCommand(containerClient, nil, "describe", "testsource")
	assert.ErrorContains(t, err, "testsource")
	assert.Assert(t, util.ContainsAll(out, "Usage", "testsource"))

	containerRecorder.Validate()
}

func TestSimpleDescribe(t *testing.T) {
	containerClient := v1alpha2.NewMockKnContainerSourceClient(t, "mynamespace")

	containerRecorder := containerClient.Recorder()
	sampleSource := createContainerSource("testsource", "docker.io/test/testimg", createSinkv1("testsvc", "default"))
	sampleSource.Namespace = "mynamespace"
	containerRecorder.GetContainerSource("testsource", sampleSource, nil)

	out, err := executeContainerSourceCommand(containerClient, nil, "describe", "testsource")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(out, "testsource", "docker.io/test/testimg", "testsvc"))
	assert.Assert(t, util.ContainsNone(out, "URI"))

	containerRecorder.Validate()
}
