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

	"gotest.tools/assert"

	"knative.dev/client/pkg/sources/v1alpha2"
	"knative.dev/client/pkg/util"
)

func TestContainerSourceDelete(t *testing.T) {

	containerClient := v1alpha2.NewMockKnContainerSourceClient(t, "testns")
	containerRecorder := containerClient.Recorder()

	containerRecorder.DeleteContainerSource("testsource", nil)

	out, err := executeContainerSourceCommand(containerClient, nil, "delete", "testsource")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(out, "deleted", "default", "testsource"))

	containerRecorder.Validate()
}

func TestDeleteWithError(t *testing.T) {

	containerClient := v1alpha2.NewMockKnContainerSourceClient(t, "mynamespace")
	containerRecorder := containerClient.Recorder()

	containerRecorder.DeleteContainerSource("testsource", errors.New("container source testsource not found"))

	out, err := executeContainerSourceCommand(containerClient, nil, "delete", "testsource")
	assert.ErrorContains(t, err, "testsource")
	assert.Assert(t, util.ContainsAll(out, "container", "source", "testsource", "not found"))

	containerRecorder.Validate()
}
