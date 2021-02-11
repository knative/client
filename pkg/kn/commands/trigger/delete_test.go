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

package trigger

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"

	eventingclientv1beta1 "knative.dev/client/pkg/eventing/v1beta1"
	"knative.dev/client/pkg/util"
)

func TestTriggerDelete(t *testing.T) {
	triggerName := "trigger-12345"

	eventingClient := eventingclientv1beta1.NewMockKnEventingClient(t)
	eventingRecorder := eventingClient.Recorder()
	eventingRecorder.DeleteTrigger(triggerName, nil)

	out, err := executeTriggerCommand(eventingClient, nil, "delete", triggerName)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(out, "deleted", "default", triggerName))

	eventingRecorder.Validate()
}

func TestTriggerDeleteWithError(t *testing.T) {
	triggerName := "trigger-12345"

	eventingClient := eventingclientv1beta1.NewMockKnEventingClient(t)
	eventingRecorder := eventingClient.Recorder()
	eventingRecorder.DeleteTrigger(triggerName, fmt.Errorf("trigger %s not found", triggerName))

	out, err := executeTriggerCommand(eventingClient, nil, "delete", triggerName)
	assert.ErrorContains(t, err, triggerName)
	assert.Assert(t, util.ContainsAll(out, "trigger", triggerName, "not found"))

	eventingRecorder.Validate()
}
