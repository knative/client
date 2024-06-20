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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"knative.dev/client/pkg/util"
)

// ResourceOptions to hold the container resource requirements values
type ResourceOptions struct {
	Requests             []string
	Limits               []string
	ResourceRequirements corev1.ResourceRequirements
}

// Validate parses the limits and requests parameters if specified and
// sets ResourceRequirements for ResourceOptions or returns error if any
func (o *ResourceOptions) Validate() ([]string, []string, error) {
	requests, requestsToRemove, err := populateResourceListV1(o.Requests)
	if err != nil {
		return []string{}, []string{}, err
	}
	o.ResourceRequirements.Requests = requests

	limits, limitsToRemove, err := populateResourceListV1(o.Limits)
	if err != nil {
		return []string{}, []string{}, err
	}
	o.ResourceRequirements.Limits = limits

	return requestsToRemove, limitsToRemove, nil
}

// populateResourceListV1 takes array of strings of form <resourceName1>=<value1>
// and returns ResourceList , an array of resource keys to remove and error if any
func populateResourceListV1(resourceStatements []string) (corev1.ResourceList, []string, error) {
	// empty input gets a nil response to preserve generator test expected behaviors
	if len(resourceStatements) == 0 {
		return nil, []string{}, nil
	}

	result := corev1.ResourceList{}
	resources, err := util.MapFromArrayAllowingSingles(resourceStatements, "=")
	if err != nil {
		return result, []string{}, err
	}

	resourcesToRemove := util.ParseMinusSuffix(resources)

	for res, value := range resources {
		parse := true
		// do not parse the quantity OR throw error if the key is being asked for removal
		for _, toRemove := range resourcesToRemove {
			if res == toRemove {
				parse = false
				break
			}
		}
		if !parse {
			continue
		}

		resourceQuantity, err := resource.ParseQuantity(value)
		if err != nil {
			return nil, resourcesToRemove, err
		}

		result[corev1.ResourceName(res)] = resourceQuantity
	}

	return result, resourcesToRemove, nil
}
