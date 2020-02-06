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

	"gotest.tools/assert"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	clientservingv1 "knative.dev/client/pkg/serving/v1"
	"knative.dev/client/pkg/util"
	"knative.dev/client/pkg/util/mock"
)

func TestServiceListAllNamespaceMock(t *testing.T) {
	client := clientservingv1.NewMockKnServiceClient(t, "default")
	r := client.Recorder()
	setupListExpectations(r)

	output, err := executeServiceCommand(client, "list", "--all-namespaces")
	assert.NilError(t, err)

	outputLines := strings.Split(output, "\n")
	assert.Assert(t, util.ContainsAll(outputLines[0], "NAMESPACE", "NAME", "URL", "LATEST", "AGE", "CONDITIONS", "READY", "REASON"))
	assert.Assert(t, util.ContainsAll(outputLines[1], "bar", "svc3"))
	assert.Assert(t, util.ContainsAll(outputLines[2], "default", "svc1"))
	assert.Assert(t, util.ContainsAll(outputLines[3], "foo", "svc2"))

	r.Validate()
}

func setupListExpectations(r *clientservingv1.ServingRecorder) {
	r.ListServices(mock.Any(), &servingv1.ServiceList{
		Items: []servingv1.Service{
			*getServiceWithNamespace("svc1", "default"),
			*getServiceWithNamespace("svc2", "foo"),
			*getServiceWithNamespace("svc3", "bar"),
		},
	}, nil)
}

func TestListEmptyMock(t *testing.T) {
	// New mock client
	client := clientservingv1.NewMockKnServiceClient(t)

	// Recording:
	r := client.Recorder()

	r.ListServices(mock.Any(), &servingv1.ServiceList{}, nil)

	output, err := executeServiceCommand(client, "list")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "No", "services", "found"))

	r.Validate()
}

func TestListEmptyWithArgMock(t *testing.T) {
	// New mock client
	client := clientservingv1.NewMockKnServiceClient(t)

	// Recording:
	r := client.Recorder()

	r.ListServices(mock.Any(), &servingv1.ServiceList{}, nil)

	output, err := executeServiceCommand(client, "list", "bar")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "No", "services", "found"))

	r.Validate()
}

func TestServiceListDefaultOutputMock(t *testing.T) {

	// New mock client
	client := clientservingv1.NewMockKnServiceClient(t)

	// Recording:
	r := client.Recorder()

	service1 := createMockServiceWithParams("foo", "default", "http://foo.default.example.com", "foo-xyz")
	service3 := createMockServiceWithParams("sss", "default", "http://sss.default.example.com", "sss-xyz")
	service2 := createMockServiceWithParams("bar", "default", "http://bar.default.example.com", "bar-xyz")
	serviceList := &servingv1.ServiceList{Items: []servingv1.Service{*service1, *service2, *service3}}
	r.ListServices(mock.Any(), serviceList, nil)

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
	client := clientservingv1.NewMockKnServiceClient(t)

	// Recording:
	r := client.Recorder()

	service1 := createMockServiceWithParams("foo", "default", "http://foo.default.example.com", "foo-xyz")
	service2 := createMockServiceWithParams("bar", "default", "http://bar.default.example.com", "bar-xyz")
	serviceList := &servingv1.ServiceList{Items: []servingv1.Service{*service1, *service2}}
	r.ListServices(mock.Any(), serviceList, nil)

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
	client := clientservingv1.NewMockKnServiceClient(t)

	// Recording:
	r := client.Recorder()

	service := createMockServiceWithParams("foo", "default", "foo.default.example.com", "foo-xyz")
	serviceList := &servingv1.ServiceList{Items: []servingv1.Service{*service}}
	r.ListServices(mock.Any(), serviceList, nil)

	output, err := executeServiceCommand(client, "list", "foo")
	assert.NilError(t, err)

	outputLines := strings.Split(output, "\n")
	assert.Check(t, util.ContainsAll(outputLines[0], "NAME", "URL", "LATEST", "AGE", "CONDITIONS", "READY", "REASON"))
	assert.Check(t, util.ContainsAll(outputLines[1], "foo", "foo.default.example.com", "foo-xyz"))

	r.Validate()
}

func TestServiceListWithTwoSrvNameMock(t *testing.T) {
	// New mock client
	client := clientservingv1.NewMockKnServiceClient(t)

	// Recording:
	r := client.Recorder()

	_, err := executeServiceCommand(client, "list", "foo", "bar")
	assert.ErrorContains(t, err, "'kn service list' accepts maximum 1 argument")

	r.Validate()
}

func getServiceWithNamespace(name, namespace string) *servingv1.Service {
	service := servingv1.Service{}
	service.Name = name
	service.Namespace = namespace
	return &service
}
