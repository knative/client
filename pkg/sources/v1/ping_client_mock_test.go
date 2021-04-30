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

	"knative.dev/eventing/pkg/apis/sources/v1"
)

func TestMockKnPingSourceClient(t *testing.T) {

	client := NewMockKnPingSourceClient(t)

	recorder := client.Recorder()

	// Record all services
	recorder.GetPingSource("hello", nil, nil)
	recorder.CreatePingSource(&v1.PingSource{}, nil)
	recorder.UpdatePingSource(&v1.PingSource{}, nil)
	recorder.DeletePingSource("hello", nil)

	// Call all service
	ctx := context.Background()
	client.GetPingSource(ctx, "hello")
	client.CreatePingSource(ctx, &v1.PingSource{})
	client.UpdatePingSource(ctx, &v1.PingSource{})
	client.DeletePingSource(ctx, "hello")

	// Validate
	recorder.Validate()
}
