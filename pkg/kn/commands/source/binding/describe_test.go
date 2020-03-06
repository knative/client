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

package binding

import (
	"errors"
	"strings"
	"testing"

	"gotest.tools/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1alpha2 "knative.dev/eventing/pkg/apis/sources/v1alpha2"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/apis/duck/v1alpha1"
	"knative.dev/pkg/tracker"

	clientv1alpha2 "knative.dev/client/pkg/sources/v1alpha2"
	"knative.dev/client/pkg/util"
)

func TestSimpleDescribeWitName(t *testing.T) {
	bindingClient := clientv1alpha2.NewMockKnSinkBindingClient(t, "mynamespace")

	bindingRecorder := bindingClient.Recorder()
	bindingRecorder.GetSinkBinding("mybinding", getSinkBindingSource("myapp", map[string]string{"foo": "bar"}), nil)

	out, err := executeSinkBindingCommand(bindingClient, nil, "describe", "mybinding")
	assert.NilError(t, err)
	util.ContainsAll(out, "mybinding", "myapp", "Deployment", "app/v1", "mynamespace", "mysvc", "foo", "bar")

	bindingRecorder.Validate()
}

func TestSimpleDescribeWithSelector(t *testing.T) {
	bindingClient := clientv1alpha2.NewMockKnSinkBindingClient(t, "mynamespace")

	bindingRecorder := bindingClient.Recorder()
	bindingRecorder.GetSinkBinding("mybinding", getSinkBindingSource("app=myapp,type=test", nil), nil)

	out, err := executeSinkBindingCommand(bindingClient, nil, "describe", "mybinding")
	assert.NilError(t, err)
	util.ContainsAll(out, "mybinding", "app:", "myapp", "type:", "test", "Deployment", "app/v1", "mynamespace", "mysvc")

	bindingRecorder.Validate()
}

func TestDescribeError(t *testing.T) {
	bindingClient := clientv1alpha2.NewMockKnSinkBindingClient(t, "mynamespace")

	bindingRecorder := bindingClient.Recorder()
	bindingRecorder.GetSinkBinding("mybinding", nil, errors.New("no sink binding mybinding found"))

	out, err := executeSinkBindingCommand(bindingClient, nil, "describe", "mybinding")
	assert.ErrorContains(t, err, "mybinding")
	util.ContainsAll(out, "mybinding")

	bindingRecorder.Validate()
}

func getSinkBindingSource(nameOrSelector string, ceOverrides map[string]string) *v1alpha2.SinkBinding {
	binding := &v1alpha2.SinkBinding{
		TypeMeta: v1.TypeMeta{},
		ObjectMeta: v1.ObjectMeta{
			Name: "mysinkbinding",
		},
		Spec: v1alpha2.SinkBindingSpec{
			SourceSpec: duckv1.SourceSpec{
				Sink: duckv1.Destination{
					Ref: &duckv1.KReference{
						Kind:      "Service",
						Namespace: "myservicenamespace",
						Name:      "mysvc",
					},
				},
			},
			BindingSpec: v1alpha1.BindingSpec{
				Subject: tracker.Reference{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Namespace:  "mynamespace",
				},
			},
		},
		Status: v1alpha2.SinkBindingStatus{},
	}

	if strings.Contains(nameOrSelector, "=") {
		selector, _ := parseSelector(nameOrSelector)
		binding.Spec.Subject.Selector = &v1.LabelSelector{
			MatchLabels: selector,
		}
	} else {
		binding.Spec.Subject.Name = nameOrSelector
	}

	if ceOverrides != nil {
		binding.Spec.CloudEventOverrides = &duckv1.CloudEventOverrides{Extensions: ceOverrides}
	}
	return binding
}
