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

package v1

import (
	"testing"
	"time"

	"knative.dev/serving/pkg/apis/serving"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"knative.dev/client/pkg/util/mock"
	"knative.dev/client/pkg/wait"
)

func TestMockKnClient(t *testing.T) {

	client := NewMockKnServiceClient(t)

	recorder := client.Recorder()

	// Record all services
	recorder.GetService("hello", nil, nil)
	recorder.ListServices(mock.Any(), nil, nil)
	recorder.CreateService(&servingv1.Service{}, nil)
	recorder.UpdateService(&servingv1.Service{}, nil)
	recorder.DeleteService("hello", time.Duration(10)*time.Second, nil)
	recorder.WaitForService("hello", time.Duration(10)*time.Second, wait.NoopMessageCallback(), nil, 10*time.Second)
	recorder.GetRevision("hello", nil, nil)
	recorder.ListRevisions(mock.Any(), nil, nil)
	recorder.DeleteRevision("hello", time.Duration(10)*time.Second, nil)
	recorder.GetRoute("hello", nil, nil)
	recorder.ListRoutes(mock.Any(), nil, nil)
	recorder.GetConfiguration("hello", nil, nil)

	// Call all services
	client.GetService("hello")
	client.ListServices(WithName("blub"))
	client.CreateService(&servingv1.Service{})
	client.UpdateService(&servingv1.Service{})
	client.DeleteService("hello", time.Duration(10)*time.Second)
	client.WaitForService("hello", time.Duration(10)*time.Second, wait.NoopMessageCallback())
	client.GetRevision("hello")
	client.ListRevisions(WithName("blub"))
	client.DeleteRevision("hello", time.Duration(10)*time.Second)
	client.GetRoute("hello")
	client.ListRoutes(WithName("blub"))
	client.GetConfiguration("hello")

	// Validate
	recorder.Validate()
}

func TestHasLabelSelector(t *testing.T) {
	assertFunction := HasLabelSelector(serving.ServiceLabelKey, "myservice")
	listConfig := []ListConfig{
		WithService("myservice"),
	}
	assertFunction(t, listConfig)
}

func TestHasFieldSelector(t *testing.T) {
	assertFunction := HasFieldSelector("metadata.name", "myname")
	listConfig := []ListConfig{
		WithName("myname"),
	}
	assertFunction(t, listConfig)
}

func TestHasSelector(t *testing.T) {
	assertFunction := HasSelector(
		[]string{serving.ServiceLabelKey, "myservice"},
		[]string{"metadata.name", "myname"})
	listConfig := []ListConfig{
		func(lo *listConfigCollector) {
			lo.Labels[serving.ServiceLabelKey] = "myservice"
			lo.Fields["metadata.name"] = "myname"
		},
	}
	assertFunction(t, listConfig)
}
