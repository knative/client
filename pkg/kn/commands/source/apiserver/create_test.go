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
	"errors"
	"testing"

	"gotest.tools/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	serving_v1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"

	knsources_v1alpha1 "knative.dev/client/pkg/eventing/sources/v1alpha1"
	knserving_client "knative.dev/client/pkg/serving/v1alpha1"
	"knative.dev/client/pkg/util"
)

func TestCreateApiServerSource(t *testing.T) {

	apiServerClient := knsources_v1alpha1.NewMockKnApiServerSourceClient(t)
	servingClient := knserving_client.NewMockKnServiceClient(t)

	servingRecorder := servingClient.Recorder()
	servingRecorder.GetService("testsvc", &serving_v1alpha1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service"},
		ObjectMeta: metav1.ObjectMeta{Name: "testsvc"},
	}, nil)

	apiServerRecorder := apiServerClient.Recorder()
	apiServerRecorder.CreateApiServerSource(createApiServerSource("testsource", "Event", "v1", "testsa", "Ref", "testsvc", false), nil)

	out, err := executeApiServerSourceCommand(apiServerClient, servingClient, "create", "testsource", "--resource", "Event:v1:false", "--service-account", "testsa", "--sink", "svc:testsvc", "--mode", "Ref")
	assert.NilError(t, err, "ApiServer source should be created")
	util.ContainsAll(out, "created", "default", "testsource")

	apiServerRecorder.Validate()
	servingRecorder.Validate()
}

func TestNoSinkError(t *testing.T) {
	servingClient := knserving_client.NewMockKnServiceClient(t)
	apiServerClient := knsources_v1alpha1.NewMockKnApiServerSourceClient(t)

	errorMsg := "cannot create ApiServerSource 'testsource' in namespace 'default' because no Service svc found"
	servingRecorder := servingClient.Recorder()
	servingRecorder.GetService("testsvc", nil, errors.New("no Service svc found"))

	out, err := executeApiServerSourceCommand(apiServerClient, servingClient, "create", "testsource", "--resource", "Event:v1:false", "--service-account", "testsa", "--sink", "svc:testsvc", "--mode", "Ref")
	assert.Error(t, err, errorMsg)
	assert.Assert(t, util.ContainsAll(out, errorMsg, "Usage"))
	servingRecorder.Validate()
}
