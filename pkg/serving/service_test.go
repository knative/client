// Copyright © 2018 The Knative Authors
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

package serving

import (
	"math/rand"
	"testing"

	"gotest.tools/assert"

	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
)

type generateNameTest struct {
	templ  string
	result string
	err    string
}

func TestGenerateName(t *testing.T) {
	rand.Seed(1)
	someRandomChars := (&revisionTemplContext{}).Random(20)
	service := &servingv1alpha1.Service{}
	service.Name = "foo"
	service.Generation = 3
	cases := []generateNameTest{
		{"{{.Service}}-v-{{.Generation}}", "foo-v-4", ""},
		{"foo-asdf", "foo-asdf", ""},
		{"{{.Bad}}", "", "can't evaluate field Bad"},
		{"{{.Service}}-{{.Random 5}}", "foo-" + someRandomChars[0:5], ""},
		{"", "", ""},
		{"andrew", "foo-andrew", ""},
		{"{{.Random 5}}", "foo-" + someRandomChars[0:5], ""},
	}
	for _, c := range cases {
		rand.Seed(1)
		name, err := GenerateRevisionName(c.templ, service)
		if c.err != "" {
			assert.ErrorContains(t, err, c.err)
		} else {
			assert.Equal(t, name, c.result)
		}
	}
}

func TestRevisionTemplateOfServiceNewStyle(t *testing.T) {
	service := &servingv1alpha1.Service{}
	service.Name = "foo"
	template := &servingv1alpha1.RevisionTemplateSpec{}
	service.Spec.Template = template
	got, err := RevisionTemplateOfService(service)
	assert.NilError(t, err)
	assert.Equal(t, got, template)
}

func TestRevisionTemplateOfServiceOldStyle(t *testing.T) {
	service := &servingv1alpha1.Service{}
	service.Name = "foo"
	template := &servingv1alpha1.RevisionTemplateSpec{}
	service.Spec.DeprecatedRunLatest = &servingv1alpha1.RunLatestType{}
	service.Spec.DeprecatedRunLatest.Configuration.DeprecatedRevisionTemplate = template
	got, err := RevisionTemplateOfService(service)
	assert.NilError(t, err)
	assert.Equal(t, got, template)
}

func TestRevisionTemplateOfServiceError(t *testing.T) {
	service := &servingv1alpha1.Service{}
	_, err := RevisionTemplateOfService(service)
	assert.ErrorContains(t, err, "does not specify")
}
