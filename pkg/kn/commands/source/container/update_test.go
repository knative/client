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

	"gotest.tools/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	dynamicfake "knative.dev/client/pkg/dynamic/fake"
	"knative.dev/client/pkg/sources/v1alpha2"
	"knative.dev/client/pkg/util"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func TestContainerSourceUpdate(t *testing.T) {
	containerClient := v1alpha2.NewMockKnContainerSourceClient(t)
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("default", &servingv1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: "serving.knative.dev/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "svc2", Namespace: "default"},
	})

	containerRecorder := containerClient.Recorder()

	present := createContainerSource("testsource", "docker.io/test/testimg", createSinkv1("svc2", "default"))
	containerRecorder.GetContainerSource("testsource", present, nil)

	updated := createContainerSource("testsource", "docker.io/test/newimg", createSinkv1("svc2", "default"))
	containerRecorder.UpdateContainerSource(updated, nil)

	output, err := executeContainerSourceCommand(containerClient, dynamicClient, "update", "testsource", "--image", "docker.io/test/newimg")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "testsource", "updated", "default"))

	containerRecorder.Validate()
}
