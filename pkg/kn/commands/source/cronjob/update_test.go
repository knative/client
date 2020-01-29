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

package cronjob

import (
	"errors"
	"testing"

	"gotest.tools/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/eventing/pkg/apis/legacysources/v1alpha1"
	"knative.dev/pkg/apis/duck/v1beta1"

	v1alpha12 "knative.dev/client/pkg/eventing/legacysources/v1alpha1"
	"knative.dev/client/pkg/util"
)

func TestSimpleUpdate(t *testing.T) {

	cronjobClient := v1alpha12.NewMockKnCronJobSourceClient(t)

	cronJobRecorder := cronjobClient.Recorder()
	cronJobRecorder.GetCronJobSource("testsource", &v1alpha1.CronJobSource{
		TypeMeta: v1.TypeMeta{},
		ObjectMeta: v1.ObjectMeta{
			Name: "testsource",
		},
		Spec: v1alpha1.CronJobSourceSpec{
			Schedule: "1 2 3 4 5",
			Data:     "maxwell",
			Sink: &v1beta1.Destination{
				Ref: &corev1.ObjectReference{
					Kind:       "Service",
					Name:       "mysvc",
					APIVersion: "serving.knative.dev/v1alpha1",
					Namespace:  "default",
				},
			},
		},
		Status: v1alpha1.CronJobSourceStatus{},
	}, nil)
	cronJobRecorder.UpdateCronJobSource(createCronJobSource("testsource", "* * * * */3", "maxwell", "mysvc"), nil)

	out, err := executeCronJobSourceCommand(cronjobClient, nil, "update", "--schedule", "* * * * */3", "testsource")
	assert.NilError(t, err)
	util.ContainsAll(out, "updated", "default", "testsource")

	cronJobRecorder.Validate()
}

func TestUpdateError(t *testing.T) {
	cronjobClient := v1alpha12.NewMockKnCronJobSourceClient(t, "mynamespace")

	cronJobRecorder := cronjobClient.Recorder()
	cronJobRecorder.GetCronJobSource("testsource", nil, errors.New("no cronjob source testsource"))

	out, err := executeCronJobSourceCommand(cronjobClient, nil, "update", "testsource")
	assert.ErrorContains(t, err, "testsource")
	util.ContainsAll(out, "Usage", "testsource")

	cronJobRecorder.Validate()
}
