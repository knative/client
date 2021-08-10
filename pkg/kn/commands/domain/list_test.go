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
	"strings"
	"testing"

	"gotest.tools/v3/assert"

	"knative.dev/client/pkg/serving/v1alpha1"
	"knative.dev/client/pkg/util"
	servingv1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"
	"knative.dev/serving/pkg/client/clientset/versioned/scheme"
)

func TestDomainMappingList(t *testing.T) {
	client := v1alpha1.NewMockKnServiceClient(t)

	dm1 := createDomainMapping("foo1", createServiceRef("foo1", "default"), "")
	dm2 := createDomainMapping("foo2", createServiceRef("foo2", "default"), "")
	servingRecorder := client.Recorder()
	servingRecorder.ListDomainMappings(&servingv1alpha1.DomainMappingList{Items: []servingv1alpha1.DomainMapping{*dm1, *dm2}}, nil)

	out, err := executeDomainCommand(client, nil, "list")
	assert.NilError(t, err, "Domain mapping should be listed")

	outputLines := strings.Split(out, "\n")
	assert.Check(t, util.ContainsAll(outputLines[0], "NAME", "URL", "READY", "KSVC"))
	assert.Check(t, util.ContainsAll(outputLines[1], "foo1"))
	assert.Check(t, util.ContainsAll(outputLines[2], "foo2"))

	servingRecorder.Validate()
}

func TestDomainMappingListEmpty(t *testing.T) {
	client := v1alpha1.NewMockKnServiceClient(t)

	servingRecorder := client.Recorder()
	servingRecorder.ListDomainMappings(&servingv1alpha1.DomainMappingList{}, nil)

	out, err := executeDomainCommand(client, nil, "list")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(out, "No", "domain", "mapping", "found"))

	servingRecorder.Validate()
}

func TestChannelListEmptyWithOutputSet(t *testing.T) {
	client := v1alpha1.NewMockKnServiceClient(t)

	servingRecorder := client.Recorder()
	domainMappingList := &servingv1alpha1.DomainMappingList{}
	err := util.UpdateGroupVersionKindWithScheme(domainMappingList, servingv1alpha1.SchemeGroupVersion, scheme.Scheme)
	assert.NilError(t, err)
	servingRecorder.ListDomainMappings(domainMappingList, nil)

	out, err := executeDomainCommand(client, nil, "list", "-o", "json")
	assert.NilError(t, err)
	assert.Check(t, util.ContainsAll(out, "\"apiVersion\": \""+servingv1alpha1.SchemeGroupVersion.String()+"\"", "\"kind\": \"DomainMappingList\"", "\"items\": []"))
	servingRecorder.Validate()
}
