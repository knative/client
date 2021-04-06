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
	"strings"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"

	"knative.dev/client/pkg/serving/v1alpha1"
	"knative.dev/client/pkg/util"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	servingv1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"
)

func TestDomainMappingDescribe(t *testing.T) {
	client := v1alpha1.NewMockKnServiceClient(t)

	servingRecorder := client.Recorder()
	servingRecorder.GetDomainMapping("foo.bar", getDomainMapping(), nil)

	out, err := executeDomainCommand(client, nil, "describe", "foo.bar")
	assert.NilError(t, err)
	assert.Assert(t, cmp.Regexp("Name:\\s+foo.bar", out))
	assert.Assert(t, cmp.Regexp("Namespace:\\s+default", out))
	assert.Assert(t, util.ContainsAll(out, "URL:", "http://foo.bar"))
	assert.Assert(t, util.ContainsAll(out, "Conditions:", "Ready"))

	// There're 2 empty lines used in the "describe" formatting
	lineCounter := 0
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			lineCounter++
		}
	}
	assert.Equal(t, lineCounter, 2)

	servingRecorder.Validate()
}

func TestDomainMappingDescribeError(t *testing.T) {
	client := v1alpha1.NewMockKnServiceClient(t)

	servingRecorder := client.Recorder()
	servingRecorder.GetDomainMapping("foo.bar", getDomainMapping(), errors.New("domainmappings.serving.knative.dev 'foo.bar' not found"))

	_, err := executeDomainCommand(client, nil, "describe", "foo.bar")
	assert.ErrorContains(t, err, "foo", "not found")

	servingRecorder.Validate()
}

func TestDomainMappingDescribeURL(t *testing.T) {
	client := v1alpha1.NewMockKnServiceClient(t)

	servingRecorder := client.Recorder()
	servingRecorder.GetDomainMapping("foo.bar", getDomainMapping(), nil)

	out, err := executeDomainCommand(client, nil, "describe", "foo.bar", "-o", "url")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(out, "http://foo.bar"))

	servingRecorder.Validate()
}

func TestDomainMappingDescribeYAML(t *testing.T) {
	client := v1alpha1.NewMockKnServiceClient(t)

	servingRecorder := client.Recorder()
	servingRecorder.GetDomainMapping("foo.bar", getDomainMapping(), nil)

	out, err := executeDomainCommand(client, nil, "describe", "foo.bar", "-o", "yaml")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(out, "kind: DomainMapping", "spec:", "status:", "metadata:"))

	servingRecorder.Validate()
}

func getDomainMapping() *servingv1alpha1.DomainMapping {
	dm := createDomainMapping("foo.bar", createServiceRef("foo", "default"))
	dm.TypeMeta = v1.TypeMeta{
		Kind:       "DomainMapping",
		APIVersion: "serving.knative.dev/v1alpha1",
	}
	dm.Status = servingv1alpha1.DomainMappingStatus{
		Status: duckv1.Status{
			Conditions: duckv1.Conditions{
				apis.Condition{
					Type:   "Ready",
					Status: "True",
				},
			},
		},
		URL: &apis.URL{Scheme: "http", Host: "foo.bar"},
	}
	return dm
}
