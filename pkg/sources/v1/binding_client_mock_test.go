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

	v1 "knative.dev/eventing/pkg/apis/sources/v1"
)

func TestMockKnClient(t *testing.T) {
	client := NewMockKnSinkBindingClient(t)

	recorder := client.Recorder()

	// Record all services
	recorder.GetSinkBinding("hello", nil, nil)
	recorder.CreateSinkBinding(&v1.SinkBinding{}, nil)
	recorder.DeleteSinkBinding("hello", nil)
	recorder.ListSinkBindings(nil, nil)
	recorder.UpdateSinkBinding(&v1.SinkBinding{}, nil)

	// Call all service
	ctx := context.Background()
	client.GetSinkBinding(ctx, "hello")
	client.CreateSinkBinding(ctx, &v1.SinkBinding{})
	client.DeleteSinkBinding(ctx, "hello")
	client.ListSinkBindings(ctx)
	client.UpdateSinkBinding(ctx, &v1.SinkBinding{})

	// Validate
	recorder.Validate()
}
