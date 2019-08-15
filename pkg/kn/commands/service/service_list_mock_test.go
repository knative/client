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
	"fmt"
	"strings"
	"testing"

	knclient "github.com/knative/client/pkg/serving/v1alpha1"
	"github.com/knative/client/pkg/util"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"gotest.tools/assert"
	"k8s.io/apimachinery/pkg/api/errors"
)

func TestServiceListAllNamespaceMock(t *testing.T) {
	client := knclient.NewMockKnClient(t)

	r := client.Recorder()

	r.GetService("svc1", nil, errors.NewNotFound(v1alpha1.Resource("service"), "svc1"))
	r.CreateService(knclient.Any(), nil)
	r.WaitForService("svc1", knclient.Any(), nil)
	r.GetService("svc1", getServiceWithNamespace("svc1", "ns1"), nil)

	r.GetService("svc2", nil, errors.NewNotFound(v1alpha1.Resource("service"), "svc2"))
	r.CreateService(knclient.Any(), nil)
	r.WaitForService("svc2", knclient.Any(), nil)
	r.GetService("svc2", getServiceWithNamespace("svc2", "ns2"), nil)

	r.GetService("svc3", nil, errors.NewNotFound(v1alpha1.Resource("service"), "svc3"))
	r.CreateService(knclient.Any(), nil)
	r.WaitForService("svc3", knclient.Any(), nil)
	r.GetService("svc3", getServiceWithNamespace("svc3", "ns3"), nil)

	r.ListServices(knclient.Any(), &v1alpha1.ServiceList{
		Items: []v1alpha1.Service{
			*getServiceWithNamespace("svc1", "ns1"),
			*getServiceWithNamespace("svc2", "ns2"),
			*getServiceWithNamespace("svc3", "ns3"),
		},
	}, nil)

	output, err := executeServiceCommand(client, "create", "svc1", "--image", "gcr.io/foo/bar:baz", "--namespace", "ns1")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "created", "svc1", "ns1", "Waiting"))

	output, err = executeServiceCommand(client, "create", "svc2", "--image", "gcr.io/foo/bar:baz", "--namespace", "ns2")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "created", "svc2", "ns2", "Waiting"))

	output, err = executeServiceCommand(client, "create", "svc3", "--image", "gcr.io/foo/bar:baz", "--namespace", "ns3")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "created", "svc3", "ns3", "Waiting"))

	output, err = executeServiceCommand(client, "list", "--all-namespaces")
	assert.NilError(t, err)

	outputLines := strings.Split(output, "\n")
	assert.Assert(t, util.ContainsAll(outputLines[0], "NAMESPACE", "NAME", "URL", "GENERATION", "AGE", "CONDITIONS", "READY", "REASON"))
	for i, line := range outputLines[1 : len(outputLines)-1] {
		ns := fmt.Sprintf("ns%d", i+1)
		svc := fmt.Sprintf("svc%d", i+1)
		assert.Check(t, util.ContainsAll(line, ns, svc))
	}

	r.Validate()
}

func getServiceWithNamespace(name, namespace string) *v1alpha1.Service {
	service := v1alpha1.Service{}
	service.Name = name
	service.Namespace = namespace
	return &service
}
