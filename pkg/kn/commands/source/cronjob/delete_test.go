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

	v1alpha12 "knative.dev/client/pkg/eventing/legacysources/v1alpha1"
	"knative.dev/client/pkg/util"
)

func TestSimpleDelete(t *testing.T) {

	cronjobClient := v1alpha12.NewMockKnCronJobSourceClient(t, "mynamespace")

	cronJobRecorder := cronjobClient.Recorder()
	cronJobRecorder.DeleteCronJobSource("testsource", nil)

	out, err := executeCronJobSourceCommand(cronjobClient, nil, "delete", "testsource")
	assert.NilError(t, err)
	util.ContainsAll(out, "deleted", "mynamespace", "testsource", "cronjob")

	cronJobRecorder.Validate()
}

func TestDeleteWithError(t *testing.T) {

	cronjobClient := v1alpha12.NewMockKnCronJobSourceClient(t, "mynamespace")

	cronJobRecorder := cronjobClient.Recorder()
	cronJobRecorder.DeleteCronJobSource("testsource", errors.New("no such cronjob source testsource"))

	out, err := executeCronJobSourceCommand(cronjobClient, nil, "delete", "testsource")
	assert.ErrorContains(t, err, "testsource")
	util.ContainsAll(out, "Usage", "no such", "testsource")

	cronJobRecorder.Validate()
}
