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
	"knative.dev/client/pkg/serving/v1beta1"
	"knative.dev/client/pkg/util"
)

func TestDomainMappingCreate(t *testing.T) {
	client := v1beta1.NewMockKnServiceClient(t)
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient(client.Namespace(), createService("foo"))

	servingRecorder := client.Recorder()
	servingRecorder.CreateDomainMapping(createDomainMapping("foo.bar", createServiceRef("foo", "default"), ""), nil)

	out, err := executeDomainCommand(client, dynamicClient, "create", "foo.bar", "--ref", "foo")
	assert.NilError(t, err, "Domain mapping should be created")
	assert.Assert(t, util.ContainsAll(out, "Domain", "mapping", "foo.bar", "created", "namespace", "default"))

	servingRecorder.Validate()

	servingRecorder.CreateDomainMapping(createDomainMapping("foo.bar", createServiceRef("foo", "default"), "my-tls-secret"), nil)

	out, err = executeDomainCommand(client, dynamicClient, "create", "foo.bar", "--ref", "foo", "--tls", "my-tls-secret")
	assert.NilError(t, err, "Domain mapping should be created")
	assert.Assert(t, util.ContainsAll(out, "Domain", "mapping", "foo.bar", "created", "namespace", "default"))

	servingRecorder.Validate()
}
func TestDomainMappingCreateWithError(t *testing.T) {
	client := v1beta1.NewMockKnServiceClient(t)
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient(client.Namespace(), createService("foo"))

	// No call should be recorded
	servingRecorder := client.Recorder()

	_, err := executeDomainCommand(client, dynamicClient, "create", "--ref", "foo")
	assert.ErrorContains(t, err, "domain create")
	assert.Assert(t, util.ContainsAll(err.Error(), "domain create", "requires", "name", "argument"))

	_, err = executeDomainCommand(client, dynamicClient, "create", "bar")
	assert.ErrorContains(t, err, "required flag")
	assert.Assert(t, util.ContainsAll(err.Error(), "required", "flag", "not", "set"))

	_, err = executeDomainCommand(client, dynamicClient, "create", "foo.bar", "--ref", "bar")
	assert.ErrorContains(t, err, "not found")
	assert.Assert(t, util.ContainsAll(err.Error(), "services", "\"bar\"", "not", "found"))

	_, err = executeDomainCommand(client, dynamicClient, "create", "foo.bar", "--ref", "foo", "--tls", "my-TLS-secret")
	assert.ErrorContains(t, err, "invalid")
	assert.Assert(t, util.ContainsAll(err.Error(), "invalid", "name", "RFC 1123 subdomain"))

	servingRecorder.Validate()
}
