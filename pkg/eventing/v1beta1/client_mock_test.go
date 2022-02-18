// Copyright © 2022 The Knative Authors
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

	"knative.dev/eventing/pkg/apis/eventing/v1beta1"
)

func TestMockKnClient(t *testing.T) {
	client := NewMockKnEventingV1beta1Client(t, "test-ns")

	recorder := client.Recorder()

	recorder.CreateEventtype(&v1beta1.EventType{}, nil)
	recorder.GetEventtype("eventtype-name", &v1beta1.EventType{}, nil)
	recorder.DeleteEventtype("eventtype-name", nil)
	recorder.ListEventtypes(&v1beta1.EventTypeList{}, nil)

	ctx := context.Background()
	client.CreateEventtype(ctx, &v1beta1.EventType{})
	client.GetEventtype(ctx, "eventtype-name")
	client.DeleteEventtype(ctx, "eventtype-name")
	client.ListEventtypes(ctx)

	recorder.Validate()
}
