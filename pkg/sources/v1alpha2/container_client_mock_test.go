/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha2

import (
	context2 "context"
	"testing"

	v1alpha2 "knative.dev/eventing/pkg/apis/sources/v1alpha2"
)

func TestMockKnConatinerSourceClient(t *testing.T) {
	client := NewMockKnContainerSourceClient(t)

	recorder := client.Recorder()

	// Record all services
	recorder.GetContainerSource("hello", nil, nil)
	recorder.CreateContainerSource(&v1alpha2.ContainerSource{}, nil)
	recorder.DeleteContainerSource("hello", nil)
	recorder.ListContainerSources(nil, nil)
	recorder.UpdateContainerSource(&v1alpha2.ContainerSource{}, nil)

	// Call all service
	client.GetContainerSource(context2.TODO(), "hello")
	client.CreateContainerSource(context2.TODO(), &v1alpha2.ContainerSource{})
	client.DeleteContainerSource("hello", context2.TODO())
	client.ListContainerSources(context2.TODO())
	client.UpdateContainerSource(context2.TODO(), &v1alpha2.ContainerSource{})

	// Validate
	recorder.Validate()
}
