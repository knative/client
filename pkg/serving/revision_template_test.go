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

package serving

import (
	"testing"

	"gotest.tools/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"knative.dev/serving/pkg/apis/autoscaling"
)

type scalingInfoTest struct {
	min    string
	max    string
	minRes int
	maxRes int
	e      bool
}

func TestScalingInfo(t *testing.T) {
	sentinel := -0xdead
	tests := []scalingInfoTest{
		{"3", "4", 3, 4, false},
		{"", "5", sentinel, 5, false},
		{"4", "", 4, sentinel, false},
		{"", "", sentinel, sentinel, false},
		{"", "funtimes", sentinel, sentinel, true},
	}
	for _, c := range tests {
		m := metav1.ObjectMeta{}
		m.Annotations = map[string]string{}
		if c.min != "" {
			m.Annotations[autoscaling.MinScaleAnnotationKey] = c.min
		}
		if c.max != "" {
			m.Annotations[autoscaling.MaxScaleAnnotationKey] = c.max
		}
		s, err := ScalingInfo(&m)
		if c.e {
			assert.Assert(t, err != nil)
			continue
		} else {
			assert.NilError(t, err)
		}
		if c.minRes != sentinel {
			assert.Assert(t, s.Min != nil)
			assert.Equal(t, c.minRes, *s.Min)
		} else {
			assert.Assert(t, s.Min == nil)
		}
		if c.maxRes != sentinel {
			assert.Assert(t, s.Max != nil)
			assert.Equal(t, c.maxRes, *s.Max)
		} else {
			assert.Assert(t, s.Max == nil)
		}

	}
}
