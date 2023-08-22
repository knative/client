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
	"errors"
	"testing"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"gotest.tools/v3/assert"

	dynamicfake "knative.dev/client/pkg/dynamic/fake"
	"knative.dev/client/pkg/serving/v1beta1"
	"knative.dev/client/pkg/util"
)

func TestDomainMappingUpdate(t *testing.T) {
	client := v1beta1.NewMockKnServiceClient(t)
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient(client.Namespace(), createService("foo"), createService("bar"))

	servingRecorder := client.Recorder()
	servingRecorder.GetDomainMapping("foo.bar", createDomainMapping("foo.bar", createServiceRef("foo", "default"), ""), nil)
	servingRecorder.UpdateDomainMapping(createDomainMapping("foo.bar", createServiceRef("bar", "default"), ""), nil)

	out, err := executeDomainCommand(client, dynamicClient, "update", "foo.bar", "--ref", "bar")
	assert.NilError(t, err, "Domain mapping should be updated")
	assert.Assert(t, util.ContainsAll(out, "Domain", "mapping", "foo.bar", "updated", "namespace", "default"))

	servingRecorder.Validate()
}

func TestDomainMappingUpdateNotFound(t *testing.T) {
	client := v1beta1.NewMockKnServiceClient(t)

	servingRecorder := client.Recorder()
	servingRecorder.GetDomainMapping("foo.bar", nil, errors.New("domainmappings.serving.knative.dev \"foo.bar\" not found"))

	_, err := executeDomainCommand(client, nil, "update", "foo.bar", "--ref", "bar")
	assert.ErrorContains(t, err, "not found")
	assert.Assert(t, util.ContainsAll(err.Error(), "domainmappings.serving.knative.dev", "\"foo.bar\"", "not", "found"))

	servingRecorder.Validate()
}

func TestDomainMappingUpdateDeletingError(t *testing.T) {
	client := v1beta1.NewMockKnServiceClient(t)

	deletingDM := createDomainMapping("foo.bar", createServiceRef("foo", "default"), "")
	deletingDM.DeletionTimestamp = &v1.Time{Time: time.Now()}

	servingRecorder := client.Recorder()
	servingRecorder.GetDomainMapping("foo.bar", deletingDM, nil)

	_, err := executeDomainCommand(client, nil, "update", "foo.bar", "--ref", "bar")
	assert.ErrorContains(t, err, "deletion")
	assert.Assert(t, util.ContainsAll(err.Error(), "can't", "update", "domain", "mapping", "foo.bar", "marked", "deletion"))

	servingRecorder.Validate()
}
