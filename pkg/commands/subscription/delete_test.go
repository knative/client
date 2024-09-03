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

package subscription

import (
	"errors"
	"testing"

	"gotest.tools/v3/assert"

	v1beta1 "knative.dev/client/pkg/messaging/v1"
	"knative.dev/client/pkg/util"
)

func TestDeleteSubscriptionErrorCase(t *testing.T) {
	cClient := v1beta1.NewMockKnSubscriptionsClient(t, "test")
	cRecorder := cClient.Recorder()
	_, err := executeSubscriptionCommand(cClient, nil, "delete")
	assert.Error(t, err, "'kn subscription delete' requires the subscription name as single argument")
	cRecorder.Validate()
}

func TestDeleteWithError(t *testing.T) {
	cClient := v1beta1.NewMockKnSubscriptionsClient(t, "test")
	cRecorder := cClient.Recorder()
	cRecorder.DeleteSubscription("sub0", errors.New("not found"))
	_, err := executeSubscriptionCommand(cClient, nil, "delete", "sub0")
	assert.ErrorContains(t, err, "not found")
	cRecorder.Validate()
}

func TestSubscriptionDelete(t *testing.T) {
	cClient := v1beta1.NewMockKnSubscriptionsClient(t, "test")
	cRecorder := cClient.Recorder()
	cRecorder.DeleteSubscription("sub0", nil)
	out, err := executeSubscriptionCommand(cClient, nil, "delete", "sub0")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(out, "deleted", "sub0", "test"))
	cRecorder.Validate()
}
