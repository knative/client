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
	"testing"

	"gotest.tools/assert"

	knsource_v1alpha1 "knative.dev/client/pkg/eventing/legacysources/v1alpha1"
	"knative.dev/client/pkg/util"
	"knative.dev/eventing/pkg/apis/legacysources/v1alpha1"
)

func TestListCronJobSource(t *testing.T) {
	cronjobClient := knsource_v1alpha1.NewMockKnCronJobSourceClient(t)

	cronJobRecorder := cronjobClient.Recorder()
	cJSource := createCronJobSource("testsource", "* * * * */2", "maxwell", "mysvc", "mysa", "100m", "128Mi", "200m", "256Mi")
	cJSourceList := v1alpha1.CronJobSourceList{}
	cJSourceList.Items = []v1alpha1.CronJobSource{*cJSource}

	cronJobRecorder.ListCronJobSource(&cJSourceList, nil)

	out, err := executeCronJobSourceCommand(cronjobClient, nil, "list")
	assert.NilError(t, err, "Sources should be listed")
	util.ContainsAll(out, "NAME", "SCHEDULE", "SINK", "AGE", "CONDITIONS", "READY", "REASON")
	util.ContainsAll(out, "testsource", "* * * * */2", "mysvc")

	cronJobRecorder.Validate()
}

func TestListCronJobSourceEmpty(t *testing.T) {
	cronjobClient := knsource_v1alpha1.NewMockKnCronJobSourceClient(t)

	cronJobRecorder := cronjobClient.Recorder()
	cJSourceList := v1alpha1.CronJobSourceList{}

	cronJobRecorder.ListCronJobSource(&cJSourceList, nil)

	out, err := executeCronJobSourceCommand(cronjobClient, nil, "list")
	assert.NilError(t, err, "Sources should be listed")
	util.ContainsNone(out, "NAME", "SCHEDULE", "SINK", "AGE", "CONDITIONS", "READY", "REASON")
	util.ContainsAll(out, "No", "CronJob", "source", "found")

	cronJobRecorder.Validate()
}
