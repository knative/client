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
	"time"

	"gotest.tools/assert"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	dynamicfake "knative.dev/client/pkg/dynamic/fake"
	"knative.dev/client/pkg/sources/v1alpha2"
	"knative.dev/client/pkg/util"
)

func TestApiServerSourceUpdate(t *testing.T) {
	apiServerClient := v1alpha2.NewMockKnAPIServerSourceClient(t)
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("default", &servingv1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: "serving.knative.dev/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "svc2", Namespace: "default"},
	})

	apiServerRecorder := apiServerClient.Recorder()

	present := createAPIServerSource("testsource", "Event", "v1", "testsa1", "Reference", map[string]string{"bla": "blub", "foo": "bar"}, createSinkv1("svc1", "default"))
	apiServerRecorder.GetAPIServerSource("testsource", present, nil)

	updated := createAPIServerSource("testsource", "Event", "v1", "testsa2", "Reference", map[string]string{"foo": "baz"}, createSinkv1("svc2", "default"))
	apiServerRecorder.UpdateAPIServerSource(updated, nil)

	output, err := executeAPIServerSourceCommand(apiServerClient, dynamicClient, "update", "testsource", "--service-account", "testsa2", "--sink", "ksvc:svc2", "--ce-override", "bla-", "--ce-override", "foo=baz")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "testsource", "updated", "default"))

	apiServerRecorder.Validate()
}

func TestApiServerSourceUpdateDeletionTimestampNotNil(t *testing.T) {
	apiServerClient := v1alpha2.NewMockKnAPIServerSourceClient(t)
	apiServerRecorder := apiServerClient.Recorder()

	present := createAPIServerSource("testsource", "Event", "v1", "testsa1", "Ref", nil, createSinkv1("svc1", "default"))
	present.DeletionTimestamp = &metav1.Time{Time: time.Now()}
	apiServerRecorder.GetAPIServerSource("testsource", present, nil)

	_, err := executeAPIServerSourceCommand(apiServerClient, nil, "update", "testsource", "--service-account", "testsa2", "--sink", "ksvc:svc2")
	assert.ErrorContains(t, err, present.Name)
	assert.ErrorContains(t, err, "deletion")
	assert.ErrorContains(t, err, "apiserver")
}
