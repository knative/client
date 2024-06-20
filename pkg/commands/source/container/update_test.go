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
	"time"

	"gotest.tools/v3/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	dynamicfake "knative.dev/client/pkg/dynamic/fake"
	v1 "knative.dev/client/pkg/sources/v1"
	"knative.dev/client/pkg/util"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func TestContainerSourceUpdate(t *testing.T) {
	containerClient := v1.NewMockKnContainerSourceClient(t)
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("default", &servingv1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: "serving.knative.dev/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "svc2", Namespace: "default"},
	})

	containerRecorder := containerClient.Recorder()

	present := createContainerSource("testsource", "docker.io/test/testimg", createSinkv1("svc1", "default"), nil, nil, nil)
	containerRecorder.GetContainerSource("testsource", present, nil)

	updated := createContainerSource("testsource", "docker.io/test/newimg", createSinkv1("svc2", "default"), nil, nil, nil)
	containerRecorder.UpdateContainerSource(updated, nil)

	output, err := executeContainerSourceCommand(containerClient, dynamicClient, "update", "testsource", "--image", "docker.io/test/newimg", "--sink", "ksvc:svc2")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "testsource", "updated", "default"))

	containerRecorder.Validate()
}

func TestContainerSourceUpdateSinkError(t *testing.T) {
	containerClient := v1.NewMockKnContainerSourceClient(t)
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("default")
	containerRecorder := containerClient.Recorder()
	present := createContainerSource("testsource", "docker.io/test/testimg", createSinkv1("svc2", "default"), nil, nil, nil)
	containerRecorder.GetContainerSource("testsource", present, nil)
	errorMsg := "cannot update ContainerSource 'testsource' in namespace 'default' because: services.serving.knative.dev \"testsvc\" not found"
	out, err := executeContainerSourceCommand(containerClient, dynamicClient, "update", "testsource", "--sink", "ksvc:testsvc")
	assert.Error(t, err, errorMsg)
	assert.Assert(t, util.ContainsAll(out, errorMsg, "Usage"))
}

func TestContainerUpdateErrorForNoArgs(t *testing.T) {
	containerClient := v1.NewMockKnContainerSourceClient(t, "mynamespace")
	out, err := executeContainerSourceCommand(containerClient, nil, "update")
	assert.ErrorContains(t, err, "single argument")
	assert.Assert(t, util.ContainsAll(out, "requires", "single argument"))
}

func TestContainerUpdateDeletionTimestampNotNil(t *testing.T) {
	containerClient := v1.NewMockKnContainerSourceClient(t, "mynamespace")
	present := createContainerSource("testsource", "docker.io/test/testimg", createSinkv1("svc1", "default"), nil, nil, nil)
	present.DeletionTimestamp = &metav1.Time{Time: time.Now()}
	containerRecorder := containerClient.Recorder()
	containerRecorder.GetContainerSource("testsource", present, nil)

	out, err := executeContainerSourceCommand(containerClient, nil, "update", "testsource")
	assert.ErrorContains(t, err, "can't update container source")
	assert.Assert(t, util.ContainsAll(out, "can't update container source", "marked for deletion"))
}

func TestContainerUpdatePSError(t *testing.T) {
	containerClient := v1.NewMockKnContainerSourceClient(t)
	containerRecorder := containerClient.Recorder()

	present := createContainerSource("testsource", "docker.io/test/testimg", createSinkv1("svc1", "default"), nil, nil, nil)
	containerRecorder.GetContainerSource("testsource", present, nil)

	out, err := executeContainerSourceCommand(containerClient, nil, "update", "testsource", "--mount", "123456")
	assert.ErrorContains(t, err, "cannot update ContainerSource")
	assert.Assert(t, util.ContainsAll(out, "cannot update ContainerSource", "Invalid --mount"))
}
