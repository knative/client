// Copyright © 2018 The Knative Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implie
// See the License for the specific language governing permissions and
// limitations under the License.

package printers

import (
	"time"

	"k8s.io/apimachinery/pkg/util/duration"
)

// CalculateAge calculates age of given resource using CreationTimestamp
// and current time.
func CalculateAge(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return duration.ShortHumanDuration(time.Now().Sub(t))
}
