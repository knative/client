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
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	dynamicfake "knative.dev/client/pkg/dynamic/fake"
	clientv1alpha1 "knative.dev/client/pkg/sources/v1alpha1"
	"knative.dev/client/pkg/util"
)

func TestCreateApiServerSource(t *testing.T) {
	testsvc := &servingv1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: "serving.knative.dev/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "testsvc", Namespace: "default"},
	}
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("default", testsvc)
	apiServerClient := clientv1alpha1.NewMockKnAPIServerSourceClient(t)

	apiServerRecorder := apiServerClient.Recorder()
	apiServerRecorder.CreateAPIServerSource(createAPIServerSource("testsource", "Event", "v1", "testsa", "Ref", "testsvc", false), nil)

	out, err := executeAPIServerSourceCommand(apiServerClient, dynamicClient, "create", "testsource", "--resource", "Event:v1:false", "--service-account", "testsa", "--sink", "svc:testsvc", "--mode", "Ref")
	assert.NilError(t, err, "ApiServer source should be created")
	util.ContainsAll(out, "created", "default", "testsource")

	apiServerRecorder.Validate()
}

func TestSinkNotFoundError(t *testing.T) {
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("default")
	apiServerClient := clientv1alpha1.NewMockKnAPIServerSourceClient(t)
	errorMsg := "cannot create ApiServerSource 'testsource' in namespace 'default' because: services.serving.knative.dev \"testsvc\" not found"
	out, err := executeAPIServerSourceCommand(apiServerClient, dynamicClient, "create", "testsource", "--resource", "Event:v1:false", "--service-account", "testsa", "--sink", "svc:testsvc", "--mode", "Ref")
	assert.Error(t, err, errorMsg)
	assert.Assert(t, util.ContainsAll(out, errorMsg, "Usage"))
}

func TestNoSinkError(t *testing.T) {
	apiServerClient := clientv1alpha1.NewMockKnAPIServerSourceClient(t)
	_, err := executeAPIServerSourceCommand(apiServerClient, nil, "create", "testsource", "--resource", "Event:v1:false", "--service-account", "testsa", "--mode", "Ref")
	assert.ErrorContains(t, err, "required flag(s)", "sink", "not set")
}
