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
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package revision

import (
	"strings"
	"testing"

	"gotest.tools/assert"
	"knative.dev/serving/pkg/apis/serving/v1"
	"knative.dev/serving/pkg/apis/serving/v1alpha1"

	"knative.dev/client/pkg/util"
)

func TestExtractTrafficAndTag(t *testing.T) {

	service := &v1alpha1.Service{
		Status: v1alpha1.ServiceStatus{
			RouteStatusFields: v1alpha1.RouteStatusFields{
				Traffic: []v1alpha1.TrafficTarget{
					createTarget("myv1", 10, "v1"),
					createTarget("myv2", 100, "v1"),
					createTarget("myv1", 20, "stable"),
				},
			},
		},
	}

	percent, tags := trafficAndTagsForRevision("myv1", service)

	assert.Equal(t, percent, int64(30), "expected percentage to be added up")
	assert.Check(t, util.ContainsAll(strings.Join(tags, ","), "v1", "stable"), "all tags included")

}

func createTarget(rev string, percent int64, tag string) v1alpha1.TrafficTarget {
	return v1alpha1.TrafficTarget{
		TrafficTarget: v1.TrafficTarget{
			Tag:          tag,
			RevisionName: rev,
			Percent:      &percent,
		},
	}
}
