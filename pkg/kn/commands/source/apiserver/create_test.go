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

	"gotest.tools/v3/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	dynamicfake "knative.dev/client/pkg/dynamic/fake"
	v1 "knative.dev/client/pkg/sources/v1"
	"knative.dev/client/pkg/util"
)

func TestCreateApiServerSource(t *testing.T) {
	testsvc := &servingv1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: "serving.knative.dev/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "testsvc", Namespace: "default"},
	}
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("default", testsvc)
	apiServerClient := v1.NewMockKnAPIServerSourceClient(t)

	apiServerRecorder := apiServerClient.Recorder()
	apiServerRecorder.CreateAPIServerSource(createAPIServerSource("testsource", "Event", "v1", "testsa", "Reference", map[string]string{"bla": "blub", "foo": "bar"}, createSinkv1("testsvc", "default")), nil)

	out, err := executeAPIServerSourceCommand(apiServerClient, dynamicClient, "create", "testsource", "--resource", "Event:v1", "--service-account", "testsa", "--sink", "ksvc:testsvc", "--mode", "Reference", "--ce-override", "bla=blub", "--ce-override", "foo=bar")
	assert.NilError(t, err, "ApiServer source should be created")
	assert.Assert(t, util.ContainsAll(out, "created", "default", "testsource"))

	apiServerRecorder.Validate()
}

func TestSinkNotFoundError(t *testing.T) {
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("default")
	apiServerClient := v1.NewMockKnAPIServerSourceClient(t)
	errorMsg := "cannot create ApiServerSource 'testsource' in namespace 'default' because: services.serving.knative.dev \"testsvc\" not found"
	out, err := executeAPIServerSourceCommand(apiServerClient, dynamicClient, "create", "testsource", "--resource", "Event:v1:key1=value1", "--service-account", "testsa", "--sink", "ksvc:testsvc", "--mode", "Reference")
	assert.Error(t, err, errorMsg)
	assert.Assert(t, util.ContainsAll(out, errorMsg, "Usage"))
}

func TestNoSinkError(t *testing.T) {
	apiServerClient := v1.NewMockKnAPIServerSourceClient(t)
	_, err := executeAPIServerSourceCommand(apiServerClient, nil, "create", "testsource", "--resource", "Event:v1", "--service-account", "testsa", "--mode", "Reference")
	assert.ErrorContains(t, err, "required flag(s)", "sink", "not set")
}
