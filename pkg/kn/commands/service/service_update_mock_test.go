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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	"knative.dev/serving/pkg/apis/serving/v1alpha1"

	servinglib "knative.dev/client/pkg/serving"
	knclient "knative.dev/client/pkg/serving/v1alpha1"

	"knative.dev/client/pkg/util"
)

func TestServiceUpdateEnvMock(t *testing.T) {
	client := knclient.NewMockKnClient(t)

	service := getService("foo")
	template, err := servinglib.RevisionTemplateOfService(service)
	assert.NilError(t, err)
	template.Spec.GetContainer().Env = []corev1.EnvVar{
		{Name: "a", Value: "mouse"},
		{Name: "b", Value: "cookie"},
		{Name: "empty", Value: ""},
	}
	template.Spec.GetContainer().Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{servinglib.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	updated := getService("foo")
	template, err = servinglib.RevisionTemplateOfService(updated)
	assert.NilError(t, err)
	template.Spec.GetContainer().Env = []corev1.EnvVar{
		{Name: "a", Value: "rabbit"},
		{Name: "b", Value: "cookie"},
	}
	template.Spec.GetContainer().Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{servinglib.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	r := client.Recorder()
	r.GetService("foo", nil, errors.NewNotFound(v1alpha1.Resource("service"), "foo"))
	r.CreateService(service, nil)
	r.GetService("foo", service, nil)
	r.UpdateService(updated, nil)

	output, err := executeServiceCommand(client, "create", "foo", "--image", "gcr.io/foo/bar:baz", "-e", "a=mouse", "--env", "b=cookie", "--env=empty", "--async", "--revision-name=")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "created", "foo", "default"))

	output, err = executeServiceCommand(client, "update", "foo", "-e", "a=rabbit", "--env=empty-", "--async", "--revision-name=")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "updated", "foo", "default"))

	r.Validate()
}
