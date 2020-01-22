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

package v1alpha1

import (
	"testing"

	"knative.dev/eventing/pkg/apis/legacysources/v1alpha1"
)

func TestMockKnCronJobSourceClient(t *testing.T) {

	client := NewMockKnCronJobSourceClient(t)

	recorder := client.Recorder()

	// Record all services
	recorder.GetCronJobSource("hello", nil, nil)
	recorder.CreateCronJobSource(&v1alpha1.CronJobSource{}, nil)
	recorder.UpdateCronJobSource(&v1alpha1.CronJobSource{}, nil)
	recorder.DeleteCronJobSource("hello", nil)

	// Call all service
	client.GetCronJobSource("hello")
	client.CreateCronJobSource(&v1alpha1.CronJobSource{})
	client.UpdateCronJobSource(&v1alpha1.CronJobSource{})
	client.DeleteCronJobSource("hello")

	// Validate
	recorder.Validate()
}
