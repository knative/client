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

package dynamic

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/client/pkg/util/mock"
)

func TestMockKnDynamicClient(t *testing.T) {

	client := NewMockKnDyanmicClient(t)

	recorder := client.Recorder()

	// Record all services
	recorder.ListCRDs(mock.Any(), nil, nil)
	recorder.ListSourcesTypes(nil, nil)
	recorder.ListSources(mock.Any(), nil, nil)

	// Call all service
	client.ListCRDs(metav1.ListOptions{})
	client.ListSourcesTypes()
	client.ListSources(WithTypeFilter("blub"))
	// Validate
	recorder.Validate()
}
