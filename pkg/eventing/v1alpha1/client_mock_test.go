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

package v1alpha1

import (
	"testing"

	"knative.dev/eventing/pkg/apis/eventing/v1alpha1"
)

func TestMockKnClient(t *testing.T) {

	client := NewMockKnEventingClient(t)

	recorder := client.Recorder()

	// Record all services
	recorder.GetTrigger("hello", nil, nil)
	recorder.CreateTrigger(&v1alpha1.Trigger{}, nil)
	recorder.DeleteTrigger("hello", nil)
	recorder.ListTriggers(nil, nil)
	recorder.UpdateTrigger(&v1alpha1.Trigger{}, nil)

	// Call all service
	client.GetTrigger("hello")
	client.CreateTrigger(&v1alpha1.Trigger{})
	client.DeleteTrigger("hello")
	client.ListTriggers()
	client.UpdateTrigger(&v1alpha1.Trigger{})

	// Validate
	recorder.Validate()
}
