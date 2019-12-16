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

package apiserver

import (
	"testing"

	"gotest.tools/assert"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	serving_v1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"

	knsources_v1alpha1 "knative.dev/client/pkg/eventing/sources/v1alpha1"
	knserving_client "knative.dev/client/pkg/serving/v1alpha1"
	"knative.dev/client/pkg/util"
)

func TestApiServerSourceUpdate(t *testing.T) {
	apiServerClient := knsources_v1alpha1.NewMockKnAPIServerSourceClient(t)
	servingClient := knserving_client.NewMockKnServiceClient(t)

	apiServerRecorder := apiServerClient.Recorder()
	servingRecorder := servingClient.Recorder()

	present := createAPIServerSource("testsource", "Event", "v1", "testsa1", "Ref", "svc1", false)
	apiServerRecorder.GetAPIServerSource("testsource", present, nil)

	servingRecorder.GetService("svc2", &serving_v1alpha1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service"},
		ObjectMeta: metav1.ObjectMeta{Name: "svc2"},
	}, nil)

	updated := createAPIServerSource("testsource", "Event", "v1", "testsa2", "Ref", "svc2", false)
	apiServerRecorder.UpdateAPIServerSource(updated, nil)

	output, err := executeAPIServerSourceCommand(apiServerClient, servingClient, "update", "testsource", "--service-account", "testsa2", "--sink", "svc:svc2")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "testsource", "updated", "default"))

	apiServerRecorder.Validate()
	servingRecorder.Validate()
}
