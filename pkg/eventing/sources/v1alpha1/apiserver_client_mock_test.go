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

	"knative.dev/eventing/pkg/apis/sources/v1alpha1"
)

func TestMockKnApiServerSourceClient(t *testing.T) {

	client := NewMockKnApiServerSourceClient(t)

	recorder := client.Recorder()

	// Record all services
	recorder.GetApiServerSource("hello", nil, nil)
	recorder.CreateApiServerSource(&v1alpha1.ApiServerSource{}, nil)
	recorder.UpdateApiServerSource(&v1alpha1.ApiServerSource{}, nil)
	recorder.DeleteApiServerSource("hello", nil)

	// Call all service
	client.GetApiServerSource("hello")
	client.CreateApiServerSource(&v1alpha1.ApiServerSource{})
	client.UpdateApiServerSource(&v1alpha1.ApiServerSource{})
	client.DeleteApiServerSource("hello")

	// Validate
	recorder.Validate()
}
