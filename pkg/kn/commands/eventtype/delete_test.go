/*
Copyright 2022 The Knative Authors

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

package eventtype

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
	"knative.dev/client/pkg/eventing/v1beta1"
	"knative.dev/client/pkg/util"
)

func TestEventtypeDelete(t *testing.T) {
	eventingClient := v1beta1.NewMockKnEventingV1beta1Client(t, testNs)

	eventingRecorder := eventingClient.Recorder()
	eventingRecorder.DeleteEventtype(eventtypeName, nil)

	out, err := executeEventtypeCommand(eventingClient, "delete", eventtypeName, "--namespace", testNs)

	assert.NilError(t, err, "Eventtype should be deleted")
	assert.Assert(t, util.ContainsAll(out, "Eventtype", eventtypeName, "successfully", "deleted", "namespace", testNs))

	eventingRecorder.Validate()
}

func TestEventtypeDeleteWithError(t *testing.T) {
	eventingClient := v1beta1.NewMockKnEventingV1beta1Client(t, "default")

	eventingRecorder := eventingClient.Recorder()
	eventingRecorder.DeleteEventtype(eventtypeName, fmt.Errorf("mock-error"))

	_, err := executeEventtypeCommand(eventingClient, "delete", eventtypeName)

	assert.ErrorContains(t, err, "cannot delete eventtype")
	assert.Assert(t, util.ContainsAll(err.Error(), "mock-error"))

	eventingRecorder.Validate()
}

func TestEventtypeDeleteWithNameMissingError(t *testing.T) {
	eventingClient := v1beta1.NewMockKnEventingV1beta1Client(t, "default")

	_, err := executeEventtypeCommand(eventingClient, "delete")

	assert.ErrorContains(t, err, "eventtype delete")
	assert.Assert(t, util.ContainsAll(err.Error(), "eventtype", "delete", "requires", "name"))
}
