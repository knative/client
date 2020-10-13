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

package traffic

import (
	"gotest.tools/assert"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"testing"

	"github.com/spf13/cobra"

	"knative.dev/client/pkg/kn/commands/flags"
)

type trafficTestCase struct {
	name             string
	existingTraffic  []servingv1.TrafficTarget
	inputFlags       []string
	desiredRevisions []string
	desiredTags      []string
	desiredPercents  []int64
}

type trafficErrorTestCase struct {
	name            string
	existingTraffic []servingv1.TrafficTarget
	inputFlags      []string
	errMsg          string
}

func newTestTrafficCommand() (*cobra.Command, *flags.Traffic) {
	var trafficFlags flags.Traffic
	trafficCmd := &cobra.Command{
		Use:   "kn",
		Short: "Traffic test kn command",
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	trafficFlags.Add(trafficCmd)
	return trafficCmd, &trafficFlags
}

func TestCompute(t *testing.T) {
	for _, testCase := range []trafficTestCase{
		{
			"assign 'latest' tag to @latest revision",
			append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("", "", 100, true)),
			[]string{"--tag", "@latest=latest"},
			[]string{"@latest"},
			[]string{"latest"},
			[]int64{100},
		},
		{
			"assign tag to revision",
			append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("", "echo-v1", 100, false)),
			[]string{"--tag", "echo-v1=stable"},
			[]string{"echo-v1"},
			[]string{"stable"},
			[]int64{100},
		},
		{
			"re-assign same tag to same revision (unchanged)",
			append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("current", "", 100, true)),
			[]string{"--tag", "@latest=current"},
			[]string{""},
			[]string{"current"},
			[]int64{100},
		},
		{
			"split traffic to tags",
			append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("", "", 100, true), newTarget("", "rev-v1", 0, false)),
			[]string{"--traffic", "@latest=10,rev-v1=90"},
			[]string{"@latest", "rev-v1"},
			[]string{"", ""},
			[]int64{10, 90},
		},
		{
			"split traffic to tags with '%' suffix",
			append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("", "", 100, true), newTarget("", "rev-v1", 0, false)),
			[]string{"--traffic", "@latest=10%,rev-v1=90%"},
			[]string{"@latest", "rev-v1"},
			[]string{"", ""},
			[]int64{10, 90},
		},
		{
			"add 2 more tagged revisions without giving them traffic portions",
			append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("latest", "", 100, true)),
			[]string{"--tag", "echo-v0=stale,echo-v1=old"},
			[]string{"@latest", "echo-v0", "echo-v1"},
			[]string{"latest", "stale", "old"},
			[]int64{100, 0, 0},
		},
		{
			"re-assign same tag to 'echo-v1' revision",
			append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("latest", "echo-v1", 100, false)),
			[]string{"--tag", "echo-v1=latest"},
			[]string{"echo-v1"},
			[]string{"latest"},
			[]int64{100},
		},
		{
			"set 2% traffic to latest revision by appending it in traffic block",
			append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("latest", "echo-v1", 100, false)),
			[]string{"--traffic", "@latest=2,echo-v1=98"},
			[]string{"echo-v1", "@latest"},
			[]string{"latest", ""},
			[]int64{98, 2},
		},
		{
			"set 2% to @latest with tag (append it in traffic block)",
			append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("latest", "echo-v1", 100, false)),
			[]string{"--traffic", "@latest=2,echo-v1=98", "--tag", "@latest=testing"},
			[]string{"echo-v1", "@latest"},
			[]string{"latest", "testing"},
			[]int64{98, 2},
		},
		{
			"change traffic percent of an existing revision in traffic block, add new revision with traffic share",
			append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("v1", "echo-v1", 100, false)),
			[]string{"--tag", "echo-v2=v2", "--traffic", "v1=10,v2=90"},
			[]string{"echo-v1", "echo-v2"},
			[]string{"v1", "v2"},
			[]int64{10, 90}, //default value,
		},
		{
			"untag 'latest' tag from 'echo-v1' revision",
			append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("latest", "echo-v1", 100, false)),
			[]string{"--untag", "latest"},
			[]string{"echo-v1"},
			[]string{""},
			[]int64{100},
		},
		{
			"replace revision pointing to 'latest' tag from 'echo-v1' to 'echo-v2' revision",
			append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("latest", "echo-v1", 50, false), newTarget("", "echo-v2", 50, false)),
			[]string{"--untag", "latest", "--tag", "echo-v1=old,echo-v2=latest"},
			[]string{"echo-v1", "echo-v2"},
			[]string{"old", "latest"},
			[]int64{50, 50},
		},
		{
			"have multiple tags for a revision, revision present in traffic block",
			append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("latest", "echo-v1", 50, false), newTarget("", "echo-v2", 50, false)),
			[]string{"--tag", "echo-v1=latest,echo-v1=current"},
			[]string{"echo-v1", "echo-v2", "echo-v1"}, // appends a new target
			[]string{"latest", "", "current"},         // with new tag requested
			[]int64{50, 50, 0},                        // and assign 0% to it
		},

		{
			"have multiple tags for a revision, revision absent in traffic block",
			append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("", "echo-v2", 100, false)),
			[]string{"--tag", "echo-v1=latest,echo-v1=current"},
			[]string{"echo-v2", "echo-v1", "echo-v1"}, // appends two new targets
			[]string{"", "latest", "current"},         // with new tags requested
			[]int64{100, 0, 0},                        // and assign 0% to each
		},
		{
			"re-assign same tag 'current' to @latest",
			append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("current", "", 100, true)),
			[]string{"--tag", "@latest=current"},
			[]string{""},
			[]string{"current"}, // since no change, no error
			[]int64{100},
		},
		{
			"assign echo-v1 10% traffic adjusting rest to @latest, echo-v1 isn't present in existing traffic block",
			append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("", "", 100, true)),
			[]string{"--traffic", "echo-v1=10,@latest=90"},
			[]string{"", "echo-v1"},
			[]string{"", ""}, // since no change, no error
			[]int64{90, 10},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if lper, lrev, ltag := len(testCase.desiredPercents), len(testCase.desiredRevisions), len(testCase.desiredTags); lper != lrev || lper != ltag {
				t.Fatalf("length of desired revisions, tags and percents is mismatched: got=(desiredPercents, desiredRevisions, desiredTags)=(%d, %d, %d)",
					lper, lrev, ltag)
			}

			testCmd, tFlags := newTestTrafficCommand()
			testCmd.SetArgs(testCase.inputFlags)
			testCmd.Execute()
			targets, err := Compute(testCmd, testCase.existingTraffic, tFlags, "serviceName")
			if err != nil {
				t.Fatal(err)
			}
			for i, target := range targets {
				if testCase.desiredRevisions[i] == "@latest" {
					assert.Equal(t, *target.LatestRevision, true)
				} else {
					assert.Equal(t, target.RevisionName, testCase.desiredRevisions[i])
				}
				assert.Equal(t, target.Tag, testCase.desiredTags[i])
				assert.Equal(t, *target.Percent, testCase.desiredPercents[i])
			}
		})
	}
}

func TestComputeErrMsg(t *testing.T) {
	for _, testCase := range []trafficErrorTestCase{
		{
			"invalid format for --traffic option",
			append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("", "", 100, true)),
			[]string{"--traffic", "@latest=100=latest"},
			"expecting the value format in value1=value2, given @latest=100=latest",
		},
		{
			"invalid format for --tag option",
			append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("", "", 100, true)),
			[]string{"--tag", "@latest="},
			"expecting the value format in value1=value2, given @latest=",
		},
		{
			"repeatedly splitting traffic to @latest revision",
			append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("", "", 100, true)),
			[]string{"--traffic", "@latest=90,@latest=10"},
			"repetition of identifier @latest is not allowed, use only once with --traffic flag",
		},
		{
			"repeatedly tagging to @latest revision not allowed",
			append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("", "", 100, true)),
			[]string{"--tag", "@latest=latest,@latest=2"},
			"repetition of identifier @latest is not allowed, use only once with --tag flag",
		},
		{
			"overwriting tag not allowed, to @latest from other revision",
			append(append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("latest", "", 2, true)), newTarget("stable", "echo-v2", 98, false)),
			[]string{"--tag", "@latest=stable"},
			"refusing to overwrite existing tag in service, add flag '--untag stable' in command to untag it",
		},
		{
			"overwriting tag not allowed, to a revision from other revision",
			append(append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("latest", "", 2, true)), newTarget("stable", "echo-v2", 98, false)),
			[]string{"--tag", "echo-v2=latest"},
			"refusing to overwrite existing tag in service, add flag '--untag latest' in command to untag it",
		},
		{
			"overwriting tag of @latest not allowed, existing != requested",
			append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("candidate", "", 100, true)),
			[]string{"--tag", "@latest=current"},
			"tag 'candidate' exists on latest ready revision of service, refusing to overwrite existing tag with 'current', add flag '--untag candidate' in command to untag it",
		},
		{
			"verify error for non integer values given to percent",
			append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("", "", 100, true)),
			[]string{"--traffic", "@latest=100p"},
			"error converting given 100p to integer value for traffic distribution",
		},
		{
			"verify error for traffic sum not equal to 100",
			append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("", "", 100, true)),
			[]string{"--traffic", "@latest=19,echo-v1=71"},
			"given traffic percents sum to 90, want 100",
		},
		{
			"verify error for values out of range given to percent",
			append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("", "", 100, true)),
			[]string{"--traffic", "@latest=-100"},
			"invalid value for traffic percent -100, expected 0 <= percent <= 100",
		},
		{
			"repeatedly splitting traffic to the same revision",
			append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("", "", 100, true)),
			[]string{"--traffic", "echo-v1=40", "--traffic", "echo-v1=60"},
			"repetition of revision reference echo-v1 is not allowed, use only once with --traffic flag",
		},
		{
			"untag single tag that does not exist",
			append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("latest", "echo-v1", 100, false)),
			[]string{"--untag", "foo"},
			"tag(s) foo not present for any revisions of service serviceName",
		},
		{
			"untag multiple tags that do not exist",
			append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("latest", "echo-v1", 100, false)),
			[]string{"--untag", "foo", "--untag", "bar"},
			"tag(s) foo, bar not present for any revisions of service serviceName",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			testCmd, tFlags := newTestTrafficCommand()
			testCmd.SetArgs(testCase.inputFlags)
			testCmd.Execute()
			_, err := Compute(testCmd, testCase.existingTraffic, tFlags, "serviceName")
			assert.Error(t, err, testCase.errMsg)
		})
	}
}
