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
	"strings"
	"testing"
	"time"

	"gotest.tools/assert"
	"k8s.io/apimachinery/pkg/api/errors"
	knclient "knative.dev/client/pkg/serving/v1alpha1"
	"knative.dev/client/pkg/util"
	"knative.dev/client/pkg/wait"

	"knative.dev/serving/pkg/apis/serving/v1alpha1"
)

func TestServiceListAllNamespaceMock(t *testing.T) {
	client := knclient.NewMockKnClient(t)

	r := client.Recorder()

	r.GetService("svc1", nil, errors.NewNotFound(v1alpha1.Resource("service"), "svc1"))
	r.CreateService(knclient.Any(), nil)
	r.WaitForService("svc1", knclient.Any(), wait.NoopMessageCallback(), nil, time.Second)
	r.GetService("svc1", getServiceWithNamespace("svc1", "default"), nil)

	r.GetService("svc2", nil, errors.NewNotFound(v1alpha1.Resource("service"), "foo"))
	r.CreateService(knclient.Any(), nil)
	r.WaitForService("svc2", knclient.Any(), wait.NoopMessageCallback(), nil, time.Second)
	r.GetService("svc2", getServiceWithNamespace("svc2", "foo"), nil)

	r.GetService("svc3", nil, errors.NewNotFound(v1alpha1.Resource("service"), "svc3"))
	r.CreateService(knclient.Any(), nil)
	r.WaitForService("svc3", knclient.Any(), wait.NoopMessageCallback(), nil, time.Second)
	r.GetService("svc3", getServiceWithNamespace("svc3", "bar"), nil)

	r.ListServices(knclient.Any(), &v1alpha1.ServiceList{
		Items: []v1alpha1.Service{
			*getServiceWithNamespace("svc1", "default"),
			*getServiceWithNamespace("svc2", "foo"),
			*getServiceWithNamespace("svc3", "bar"),
		},
	}, nil)

	output, err := executeServiceCommand(client, "create", "svc1", "--image", "gcr.io/foo/bar:baz", "--namespace", "default")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "Creating", "svc1", "default", "Ready"))

	r.Namespace("foo")
	output, err = executeServiceCommand(client, "create", "svc2", "--image", "gcr.io/foo/bar:baz", "--namespace", "foo")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "Creating", "svc2", "foo", "Ready"))

	r.Namespace("bar")
	output, err = executeServiceCommand(client, "create", "svc3", "--image", "gcr.io/foo/bar:baz", "--namespace", "bar")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "Creating", "svc3", "bar", "Ready"))

	output, err = executeServiceCommand(client, "list", "--all-namespaces")
	assert.NilError(t, err)

	outputLines := strings.Split(output, "\n")
	assert.Assert(t, util.ContainsAll(outputLines[0], "NAMESPACE", "NAME", "URL", "LATEST", "AGE", "CONDITIONS", "READY", "REASON"))
	assert.Assert(t, util.ContainsAll(outputLines[1], "default", "svc1"))
	assert.Assert(t, util.ContainsAll(outputLines[2], "bar", "svc3"))
	assert.Assert(t, util.ContainsAll(outputLines[3], "foo", "svc2"))

	r.Validate()
}

func TestListEmptyMock(t *testing.T) {
	// New mock client
	client := knclient.NewMockKnClient(t)

	// Recording:
	r := client.Recorder()

	r.ListServices(knclient.Any(), &v1alpha1.ServiceList{}, nil)

	output, err := executeServiceCommand(client, "list")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "No", "services", "found"))

	r.Validate()
}

func TestListEmptyWithArgMock(t *testing.T) {
	// New mock client
	client := knclient.NewMockKnClient(t)

	// Recording:
	r := client.Recorder()

	r.ListServices(knclient.Any(), &v1alpha1.ServiceList{}, nil)

	output, err := executeServiceCommand(client, "list", "bar")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "No", "services", "found"))

	r.Validate()
}

func TestServiceListDefaultOutputMock(t *testing.T) {

	// New mock client
	client := knclient.NewMockKnClient(t)

	// Recording:
	r := client.Recorder()

	service1 := createMockServiceWithParams("foo", "default", "http://foo.default.example.com", "foo-xyz")
	service3 := createMockServiceWithParams("sss", "default", "http://sss.default.example.com", "sss-xyz")
	service2 := createMockServiceWithParams("bar", "default", "http://bar.default.example.com", "bar-xyz")
	serviceList := &v1alpha1.ServiceList{Items: []v1alpha1.Service{*service1, *service2, *service3}}
	r.ListServices(knclient.Any(), serviceList, nil)

	output, err := executeServiceCommand(client, "list")
	assert.NilError(t, err)

	outputLines := strings.Split(output, "\n")
	assert.Check(t, util.ContainsAll(outputLines[0], "NAME", "URL", "LATEST", "AGE", "CONDITIONS", "READY", "REASON"))
	assert.Check(t, util.ContainsAll(outputLines[1], "bar", "bar.default.example.com", "bar-xyz"))
	assert.Check(t, util.ContainsAll(outputLines[2], "foo", "foo.default.example.com", "foo-xyz"))
	assert.Check(t, util.ContainsAll(outputLines[3], "sss", "sss.default.example.com", "sss-xyz"))

	r.Validate()
}

func TestServiceListDefaultOutputNoHeadersMock(t *testing.T) {
	// New mock client
	client := knclient.NewMockKnClient(t)

	// Recording:
	r := client.Recorder()

	service1 := createMockServiceWithParams("foo", "default", "http://foo.default.example.com", "foo-xyz")
	service2 := createMockServiceWithParams("bar", "default", "http://bar.default.example.com", "bar-xyz")
	serviceList := &v1alpha1.ServiceList{Items: []v1alpha1.Service{*service1, *service2}}
	r.ListServices(knclient.Any(), serviceList, nil)

	output, err := executeServiceCommand(client, "list", "--no-headers")
	assert.NilError(t, err)

	outputLines := strings.Split(output, "\n")
	assert.Check(t, util.ContainsNone(outputLines[0], "NAME", "URL", "LATEST", "AGE", "CONDITIONS", "READY", "REASON"))
	assert.Check(t, util.ContainsAll(outputLines[0], "bar", "bar.default.example.com", "bar-xyz"))
	assert.Check(t, util.ContainsAll(outputLines[1], "foo", "foo.default.example.com", "foo-xyz"))

	r.Validate()
}

func TestServiceListOneOutputMock(t *testing.T) {
	// New mock client
	client := knclient.NewMockKnClient(t)

	// Recording:
	r := client.Recorder()

	service := createMockServiceWithParams("foo", "default", "foo.default.example.com", "foo-xyz")
	serviceList := &v1alpha1.ServiceList{Items: []v1alpha1.Service{*service}}
	r.ListServices(knclient.Any(), serviceList, nil)

	output, err := executeServiceCommand(client, "list", "foo")
	assert.NilError(t, err)

	outputLines := strings.Split(output, "\n")
	assert.Check(t, util.ContainsAll(outputLines[0], "NAME", "URL", "LATEST", "AGE", "CONDITIONS", "READY", "REASON"))
	assert.Check(t, util.ContainsAll(outputLines[1], "foo", "foo.default.example.com", "foo-xyz"))

	r.Validate()
}

func TestServiceListWithTwoSrvNameMock(t *testing.T) {
	// New mock client
	client := knclient.NewMockKnClient(t)

	// Recording:
	r := client.Recorder()

	_, err := executeServiceCommand(client, "list", "foo", "bar")
	assert.ErrorContains(t, err, "'kn service list' accepts maximum 1 argument")

	r.Validate()
}

func getServiceWithNamespace(name, namespace string) *v1alpha1.Service {
	service := v1alpha1.Service{}
	service.Name = name
	service.Namespace = namespace
	return &service
}
