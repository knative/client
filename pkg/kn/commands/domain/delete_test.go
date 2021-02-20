// Copyright © 2021 The Knative Authors
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

package domain

import (
	"testing"

	"gotest.tools/v3/assert"

	dynamicfake "knative.dev/client/pkg/dynamic/fake"
	"knative.dev/client/pkg/serving/v1alpha1"
	"knative.dev/client/pkg/util"
)

func TestDomainMappingDelete(t *testing.T) {
	client := v1alpha1.NewMockKnServiceClient(t)

	servingRecorder := client.Recorder()
	servingRecorder.DeleteDomainMapping("foo.bar", nil)

	out, err := executeDomainCommand(client, nil, "delete", "foo.bar")
	assert.NilError(t, err, "Domain mapping should be deleted")
	assert.Assert(t, util.ContainsAll(out, "Domain", "mapping", "foo.bar", "deleted", "namespace", "default"))

	servingRecorder.Validate()
}
func TestDomainMappingDeleteWithError(t *testing.T) {
	client := v1alpha1.NewMockKnServiceClient(t)
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient(client.Namespace(), createService("foo"))

	// No call should be recorded
	servingRecorder := client.Recorder()

	_, err := executeDomainCommand(client, dynamicClient, "delete")
	assert.ErrorContains(t, err, "domain delete")
	assert.Assert(t, util.ContainsAll(err.Error(), "domain delete", "requires", "name", "argument"))

	servingRecorder.Validate()
}
