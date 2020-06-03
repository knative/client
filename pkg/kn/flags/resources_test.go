// Copyright © 2020 The Knative Authors
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

package flags

import (
	"testing"

	"gotest.tools/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type resourceOptionsTestCase struct {
	requests         []string
	limits           []string
	expectedRequests corev1.ResourceList
	expectedLimits   corev1.ResourceList
	expectedErr      bool
}

func parseQuantity(value string) resource.Quantity {
	q, _ := resource.ParseQuantity(value)
	return q
}

func TestResourceOptions(t *testing.T) {
	cases := []*resourceOptionsTestCase{
		{[]string{"memory=200Mi", "cpu=200m"},
			[]string{"memory=1024Mi", "cpu=500m"},
			corev1.ResourceList{corev1.ResourceMemory: parseQuantity("200Mi"),
				corev1.ResourceCPU: parseQuantity("200m")},
			corev1.ResourceList{corev1.ResourceMemory: parseQuantity("1024Mi"),
				corev1.ResourceCPU: parseQuantity("500m")},
			false,
		},
		{[]string{},
			[]string{"nvidia.com/gpu=1"},
			nil,
			corev1.ResourceList{corev1.ResourceName("nvidia.com/gpu"): parseQuantity("1")},
			false,
		},

		{[]string{},
			[]string{"memory:500Mi"},
			nil,
			nil,
			true,
		},
		{[]string{"memory:200Mi"},
			[]string{},
			nil,
			nil,
			true,
		},
		{[]string{"memory=200MB"},
			[]string{},
			nil,
			nil,
			true,
		},
		{[]string{"cpu=500m", "cpu-"},
			[]string{},
			corev1.ResourceList{},
			nil,
			false,
		},
		{[]string{},
			// resource being asked for removal is considered to be removed and its value isnt validated (200MB which is incorrect)
			[]string{"memory=200Mi", "cpu=200m", "memory-=200MB"},
			nil,
			corev1.ResourceList{corev1.ResourceName("cpu"): parseQuantity("200m")},
			false,
		},
	}
	for _, c := range cases {
		options := &ResourceOptions{}
		options.Requests = c.requests
		options.Limits = c.limits
		reqToRemove, limToRemove, err := options.Validate()

		if c.expectedErr {
			assert.Assert(t, err != nil)
		} else {
			// do the resource removal here as we arent dealing with container template in tests
			if len(reqToRemove) > 0 {
				for _, req := range reqToRemove {
					delete(options.ResourceRequirements.Requests, corev1.ResourceName(req))
				}
			}

			if len(limToRemove) > 0 {
				for _, lim := range limToRemove {
					delete(options.ResourceRequirements.Limits, corev1.ResourceName(lim))
				}
			}

			assert.DeepEqual(t, options.ResourceRequirements, corev1.ResourceRequirements{Limits: c.expectedLimits, Requests: c.expectedRequests})
		}
	}
}
