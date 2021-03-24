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

package v1beta1

import (
	"context"
	"testing"
	"time"

	v1beta1 "knative.dev/eventing/pkg/apis/eventing/v1beta1"
)

func TestMockKnClient(t *testing.T) {

	client := NewMockKnEventingClient(t)

	recorder := client.Recorder()

	// Record all services
	recorder.GetTrigger("hello", nil, nil)
	recorder.CreateTrigger(&v1beta1.Trigger{}, nil)
	recorder.DeleteTrigger("hello", nil)
	recorder.ListTriggers(nil, nil)
	recorder.UpdateTrigger(&v1beta1.Trigger{}, nil)

	recorder.CreateBroker(&v1beta1.Broker{}, nil)
	recorder.GetBroker("foo", nil, nil)
	recorder.DeleteBroker("foo", time.Duration(10)*time.Second, nil)
	recorder.ListBrokers(nil, nil)

	// Call all service
	client.GetTrigger(context.Background(), "hello")
	client.CreateTrigger(context.Background(), &v1beta1.Trigger{})
	client.DeleteTrigger(context.Background(), "hello")
	client.ListTriggers(context.Background())
	client.UpdateTrigger(context.Background(), &v1beta1.Trigger{})

	client.CreateBroker(context.Background(), &v1beta1.Broker{})
	client.GetBroker(context.Background(), "foo")
	client.DeleteBroker(context.Background(), "foo", time.Duration(10)*time.Second)
	client.ListBrokers(context.Background())

	// Validate
	recorder.Validate()
}
