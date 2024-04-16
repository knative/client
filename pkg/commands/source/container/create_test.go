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

	dynamicfake "knative.dev/client-pkg/pkg/dynamic/fake"
	v1 "knative.dev/client-pkg/pkg/sources/v1"
	"knative.dev/client-pkg/pkg/util"
)

func TestCreateContainerSource(t *testing.T) {
	testsvc := &servingv1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: "serving.knative.dev/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "testsvc", Namespace: "default"},
	}
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("default", testsvc)
	containerClient := v1.NewMockKnContainerSourceClient(t)

	containerRecorder := containerClient.Recorder()
	containerRecorder.CreateContainerSource(createContainerSource("testsource", "docker.io/test/testimg", createSinkv1("testsvc", "default"), nil, nil, nil), nil)

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

func TestContainerCreateErrorForNoArgs(t *testing.T) {
	containerClient := v1.NewMockKnContainerSourceClient(t, "mynamespace")
	argMissingMsg := "requires the name of the source to create as single argument"
	_, err := executeContainerSourceCommand(containerClient, nil, "create", "--sink", "ksvc:testsvc", "--image", "docker.io/test/testimg")
	assert.Error(t, err, argMissingMsg)
}

func TestContainerCreatePSError(t *testing.T) {
	testsvc := &servingv1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: "serving.knative.dev/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "testsvc", Namespace: "default"},
	}
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("default", testsvc)
	containerClient := v1.NewMockKnContainerSourceClient(t)

	out, err := executeContainerSourceCommand(containerClient, dynamicClient, "create", "testsource", "--sink", "ksvc:testsvc", "--image", "docker.io/test/testimg", "--mount", "123456")
	assert.ErrorContains(t, err, "cannot create ContainerSource")
	assert.Assert(t, util.ContainsAll(out, "cannot create ContainerSource", "Invalid --mount"))
}
