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
	"errors"
	"testing"

	"gotest.tools/v3/assert"

	clientservingv1 "knative.dev/client/pkg/serving/v1"
	"knative.dev/client/pkg/util"
	"knative.dev/client/pkg/util/mock"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func TestServiceDeleteMock(t *testing.T) {
	// New mock client
	client := clientservingv1.NewMockKnServiceClient(t)

	// Recording:
	r := client.Recorder()

	r.DeleteService("foo", mock.Any(), nil)

	output, err := executeServiceCommand(client, "delete", "foo")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "deleted", "foo", "default"))

	r.Validate()

}

func TestServiceDeleteMockNoWait(t *testing.T) {
	// New mock client
	client := clientservingv1.NewMockKnServiceClient(t)

	// Recording:
	r := client.Recorder()

	r.DeleteService("foo", mock.Any(), nil)

	output, err := executeServiceCommand(client, "delete", "foo", "--no-wait")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "deleted", "foo", "default"))

	r.Validate()

}

func TestMultipleServiceDeleteMock(t *testing.T) {
	// New mock client
	client := clientservingv1.NewMockKnServiceClient(t)

	// Recording:
	r := client.Recorder()
	// Wait for delete event

	r.DeleteService("foo", mock.Any(), nil)
	r.DeleteService("bar", mock.Any(), nil)
	r.DeleteService("baz", mock.Any(), nil)

	output, err := executeServiceCommand(client, "delete", "foo", "bar", "baz")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "deleted", "foo", "bar", "baz", "default"))

	r.Validate()
}

func TestServiceDeleteAllMock(t *testing.T) {
	// New mock client
	client := clientservingv1.NewMockKnServiceClient(t)

	// Recording:
	r := client.Recorder()

	// Wait for delete event
	r.DeleteService("foo", mock.Any(), nil)
	r.DeleteService("bar", mock.Any(), nil)
	r.DeleteService("baz", mock.Any(), nil)

	service1 := createMockServiceWithParams("foo", "default", "http://foo.default.example.com", "foo-xyz")
	service2 := createMockServiceWithParams("bar", "default", "http://bar.default.example.com", "bar-xyz")
	service3 := createMockServiceWithParams("baz", "default", "http://baz.default.example.com", "baz-xyz")
	serviceList := &servingv1.ServiceList{Items: []servingv1.Service{*service1, *service2, *service3}}
	r.ListServices(mock.Any(), serviceList, nil)

	output, err := executeServiceCommand(client, "delete", "--all")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "deleted", "foo", "bar", "baz", "default"))

	r.Validate()
}

func TestServiceDeleteAllErrorFromArgMock(t *testing.T) {
	// New mock client
	client := clientservingv1.NewMockKnServiceClient(t)

	_, err := executeServiceCommand(client, "delete", "foo", "--all")
	assert.Error(t, err, "'service delete' with --all flag requires no arguments")
}

func TestServiceDeleteAllNoServicesMock(t *testing.T) {
	// New mock client
	client := clientservingv1.NewMockKnServiceClient(t)

	// Recording:
	r := client.Recorder()

	serviceList := &servingv1.ServiceList{Items: []servingv1.Service{}}
	r.ListServices(mock.Any(), serviceList, nil)

	output, err := executeServiceCommand(client, "delete", "--all")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "No", "services", "found"))

	r.Validate()
}

func TestServiceDeleteNoSvcNameMock(t *testing.T) {
	// New mock client
	client := clientservingv1.NewMockKnServiceClient(t)

	// Recording:
	r := client.Recorder()

	_, err := executeServiceCommand(client, "delete")
	assert.ErrorContains(t, err, "requires the service name")

	r.Validate()

}

func TestServiceDeleteCheckErrorForNotFoundServicesMock(t *testing.T) {
	// New mock client
	client := clientservingv1.NewMockKnServiceClient(t)

	// Recording:
	r := client.Recorder()

	r.DeleteService("foo", mock.Any(), nil)
	r.DeleteService("bar", mock.Any(), errors.New("services.serving.knative.dev \"bar\" not found."))
	r.DeleteService("baz", mock.Any(), errors.New("services.serving.knative.dev \"baz\" not found."))

	output, err := executeServiceCommand(client, "delete", "foo", "bar", "baz")
	if err == nil {
		t.Fatal("Expected service not found error, returned nil")
	}
	assert.Assert(t, util.ContainsAll(output, "'foo' successfully deleted", "\"bar\" not found", "\"baz\" not found"))

	r.Validate()
}
