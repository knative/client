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

package service

import (
	"testing"

	"gotest.tools/assert"

	knclient "knative.dev/client/pkg/serving/v1alpha1"
	"knative.dev/client/pkg/util"
)

func TestServiceDeleteMock(t *testing.T) {
	// New mock client
	client := knclient.NewMockKnClient(t)

	// Recording:
	r := client.Recorder()

	r.DeleteService("foo", nil)

	output, err := executeServiceCommand(client, "delete", "foo")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "deleted", "foo", "default"))

	r.Validate()

}

func TestMultipleServiceDeleteMock(t *testing.T) {
	// New mock client
	client := knclient.NewMockKnClient(t)

	// Recording:
	r := client.Recorder()

	r.DeleteService("foo", nil)
	r.DeleteService("bar", nil)
	r.DeleteService("baz", nil)

	output, err := executeServiceCommand(client, "delete", "foo", "bar", "baz")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "deleted", "foo", "bar", "baz", "default"))

	r.Validate()
}

func TestServiceDeleteNoSvcNameMock(t *testing.T) {
	// New mock client
	client := knclient.NewMockKnClient(t)

	// Recording:
	r := client.Recorder()

	_, err := executeServiceCommand(client, "delete")
	assert.ErrorContains(t, err, "requires the service name")

	r.Validate()

}
