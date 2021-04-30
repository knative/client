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

package container

import (
	"testing"

	"gotest.tools/v3/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	dynamicfake "knative.dev/client/pkg/dynamic/fake"
	v1 "knative.dev/client/pkg/sources/v1"
	"knative.dev/client/pkg/util"
)

func TestCreateContainerSource(t *testing.T) {
	testsvc := &servingv1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: "serving.knative.dev/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "testsvc", Namespace: "default"},
	}
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("default", testsvc)
	containerClient := v1.NewMockKnContainerSourceClient(t)

	containerRecorder := containerClient.Recorder()
	containerRecorder.CreateContainerSource(createContainerSource("testsource", "docker.io/test/testimg", createSinkv1("testsvc", "default")), nil)

	out, err := executeContainerSourceCommand(containerClient, dynamicClient, "create", "testsource", "--image", "docker.io/test/testimg", "--sink", "ksvc:testsvc")
	assert.NilError(t, err, "Container source should be created")
	assert.Assert(t, util.ContainsAll(out, "created", "default", "testsource"))

	containerRecorder.Validate()
}

func TestSinkNotFoundError(t *testing.T) {
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("default")
	containerClient := v1.NewMockKnContainerSourceClient(t)
	errorMsg := "cannot create ContainerSource 'testsource' in namespace 'default' because: services.serving.knative.dev \"testsvc\" not found"
	out, err := executeContainerSourceCommand(containerClient, dynamicClient, "create", "testsource", "--image", "docker.io/test/testimg", "--sink", "ksvc:testsvc")
	assert.Error(t, err, errorMsg)
	assert.Assert(t, util.ContainsAll(out, errorMsg, "Usage"))
}

func TestNoSinkError(t *testing.T) {
	containerClient := v1.NewMockKnContainerSourceClient(t)
	_, err := executeContainerSourceCommand(containerClient, nil, "create", "testsource", "--image", "docker.io/test/testimg")
	assert.ErrorContains(t, err, "required flag(s)", "sink", "not set")
}
