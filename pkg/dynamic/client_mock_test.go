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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
	"knative.dev/client/pkg/util/mock"
)

func TestMockKnDynamicClient(t *testing.T) {

	client := NewMockKnDynamicClient(t)

	recorder := client.Recorder()

	recorder.ListCRDs(mock.Any(), nil, nil)
	recorder.ListSourcesTypes(nil, nil)
	recorder.ListSources(mock.Any(), nil, nil)
	recorder.ListChannelsTypes(nil, nil)
	recorder.RawClient(&fake.FakeDynamicClient{})
	recorder.ListSourcesUsingGVKs(mock.Any(), mock.Any(), nil, nil)
	recorder.ListChannelsUsingGVKs(mock.Any(), mock.Any(), nil, nil)

	client.ListCRDs(metav1.ListOptions{})
	client.ListSourcesTypes()
	client.ListChannelsTypes()
	client.ListSources(WithTypeFilter("blub"))
	client.RawClient()
	client.ListSourcesUsingGVKs(&[]schema.GroupVersionKind{}, WithTypeFilter("blub"))
	client.ListChannelsUsingGVKs(&[]schema.GroupVersionKind{}, WithTypeFilter("blub"))

	// Validate
	recorder.Validate()
}
