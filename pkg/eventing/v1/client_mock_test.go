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
	"context"
	"testing"
	"time"

	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
)

func TestMockKnClient(t *testing.T) {

	client := NewMockKnEventingClient(t)

	recorder := client.Recorder()

	// Record all services
	recorder.GetTrigger("hello", nil, nil)
	recorder.CreateTrigger(&eventingv1.Trigger{}, nil)
	recorder.DeleteTrigger("hello", nil)
	recorder.ListTriggers(nil, nil)
	recorder.UpdateTrigger(&eventingv1.Trigger{}, nil)
	recorder.GetTrigger("hello", &eventingv1.Trigger{}, nil)
	recorder.UpdateTrigger(&eventingv1.Trigger{}, nil)

	recorder.CreateBroker(&eventingv1.Broker{}, nil)
	recorder.GetBroker("foo", nil, nil)
	recorder.DeleteBroker("foo", time.Duration(10)*time.Second, nil)
	recorder.ListBrokers(nil, nil)

	// Call all service
	ctx := context.Background()
	client.GetTrigger(ctx, "hello")
	client.CreateTrigger(ctx, &eventingv1.Trigger{})
	client.DeleteTrigger(ctx, "hello")
	client.ListTriggers(ctx)
	client.UpdateTrigger(ctx, &eventingv1.Trigger{})
	client.UpdateTriggerWithRetry(ctx, "hello", func(origSource *eventingv1.Trigger) (*eventingv1.Trigger, error) {
		return origSource, nil
	}, 10)

	client.CreateBroker(ctx, &eventingv1.Broker{})
	client.GetBroker(ctx, "foo")
	client.DeleteBroker(ctx, "foo", time.Duration(10)*time.Second)
	client.ListBrokers(ctx)

	// Validate
	recorder.Validate()
}
