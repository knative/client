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
	"gotest.tools/v3/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/client/pkg/kn/commands/revision"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"testing"

	"github.com/spf13/cobra"

	"knative.dev/client/pkg/kn/commands/flags"
)

type trafficTestCase struct {
	name              string
	existingTraffic   []servingv1.TrafficTarget
	inputFlags        []string
	desiredRevisions  []string
	desiredTags       []string
	desiredPercents   []int64
	existingRevisions []servingv1.Revision
}

type trafficErrorTestCase struct {
	name              string
	existingTraffic   []servingv1.TrafficTarget
	inputFlags        []string
	errMsg            string
	existingRevisions []servingv1.Revision
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
			name:             "assign 'latest' tag to @latest revision",
			existingTraffic:  append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("", "", 100, true)),
			inputFlags:       []string{"--tag", "@latest=latest"},
			desiredRevisions: []string{"@latest"},
			desiredTags:      []string{"latest"},
			desiredPercents:  []int64{100},
		},
		{
			name:             "assign tag to revision",
			existingTraffic:  append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("", "echo-v1", 100, false)),
			inputFlags:       []string{"--tag", "echo-v1=stable"},
			desiredRevisions: []string{"echo-v1"},
			desiredTags:      []string{"stable"},
			desiredPercents:  []int64{100},
		},
		{
			name:             "re-assign same tag to same revision (unchanged)",
			existingTraffic:  append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("current", "", 100, true)),
			inputFlags:       []string{"--tag", "@latest=current"},
			desiredRevisions: []string{""},
			desiredTags:      []string{"current"},
			desiredPercents:  []int64{100},
		},
		{
			name:             "split traffic to tags",
			existingTraffic:  append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("", "", 100, true), newTarget("", "rev-v1", 0, false)),
			inputFlags:       []string{"--traffic", "@latest=10,rev-v1=90"},
			desiredRevisions: []string{"@latest", "rev-v1"},
			desiredTags:      []string{"", ""},
			desiredPercents:  []int64{10, 90},
		},
		{
			name:             "split traffic to tags with '%' suffix",
			existingTraffic:  append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("", "", 100, true), newTarget("", "rev-v1", 0, false)),
			inputFlags:       []string{"--traffic", "@latest=10%,rev-v1=90%"},
			desiredRevisions: []string{"@latest", "rev-v1"},
			desiredTags:      []string{"", ""},
			desiredPercents:  []int64{10, 90},
		},
		{
			name:             "add 2 more tagged revisions without giving them traffic portions",
			existingTraffic:  append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("latest", "", 100, true)),
			inputFlags:       []string{"--tag", "echo-v0=stale,echo-v1=old"},
			desiredRevisions: []string{"@latest", "echo-v0", "echo-v1"},
			desiredTags:      []string{"latest", "stale", "old"},
			desiredPercents:  []int64{100, 0, 0},
		},
		{
			name:             "re-assign same tag to 'echo-v1' revision",
			existingTraffic:  append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("latest", "echo-v1", 100, false)),
			inputFlags:       []string{"--tag", "echo-v1=latest"},
			desiredRevisions: []string{"echo-v1"},
			desiredTags:      []string{"latest"},
			desiredPercents:  []int64{100},
		},
		{
			name:             "set 2% traffic to latest revision by appending it in traffic block",
			existingTraffic:  append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("latest", "echo-v1", 100, false)),
			inputFlags:       []string{"--traffic", "@latest=2,echo-v1=98"},
			desiredRevisions: []string{"echo-v1", "@latest"},
			desiredTags:      []string{"latest", ""},
			desiredPercents:  []int64{98, 2},
		},
		{
			name:             "set 2% to @latest with tag (append it in traffic block)",
			existingTraffic:  append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("latest", "echo-v1", 100, false)),
			inputFlags:       []string{"--traffic", "@latest=2,echo-v1=98", "--tag", "@latest=testing"},
			desiredRevisions: []string{"echo-v1", "@latest"},
			desiredTags:      []string{"latest", "testing"},
			desiredPercents:  []int64{98, 2},
		},
		{
			name:             "change traffic percent of an existing revision in traffic block, add new revision with traffic share",
			existingTraffic:  append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("v1", "echo-v1", 100, false)),
			inputFlags:       []string{"--tag", "echo-v2=v2", "--traffic", "v1=10,v2=90"},
			desiredRevisions: []string{"echo-v1", "echo-v2"},
			desiredTags:      []string{"v1", "v2"},
			desiredPercents:  []int64{10, 90}, //default value,
		},
		{
			name:             "untag 'latest' tag from 'echo-v1' revision",
			existingTraffic:  append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("latest", "echo-v1", 100, false)),
			inputFlags:       []string{"--untag", "latest"},
			desiredRevisions: []string{"echo-v1"},
			desiredTags:      []string{""},
			desiredPercents:  []int64{100},
		},
		{
			name:             "replace revision pointing to 'latest' tag from 'echo-v1' to 'echo-v2' revision",
			existingTraffic:  append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("latest", "echo-v1", 50, false), newTarget("", "echo-v2", 50, false)),
			inputFlags:       []string{"--untag", "latest", "--tag", "echo-v1=old,echo-v2=latest"},
			desiredRevisions: []string{"echo-v1", "echo-v2"},
			desiredTags:      []string{"old", "latest"},
			desiredPercents:  []int64{50, 50},
		},
		{
			name:             "have multiple tags for a revision, revision present in traffic block",
			existingTraffic:  append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("latest", "echo-v1", 50, false), newTarget("", "echo-v2", 50, false)),
			inputFlags:       []string{"--tag", "echo-v1=latest,echo-v1=current"},
			desiredRevisions: []string{"echo-v1", "echo-v2", "echo-v1"}, // appends a new target
			desiredTags:      []string{"latest", "", "current"},         // with new tag requested
			desiredPercents:  []int64{50, 50, 0},                        // and assign 0% to it
		},

		{
			name:             "have multiple tags for a revision, revision absent in traffic block",
			existingTraffic:  append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("", "echo-v2", 100, false)),
			inputFlags:       []string{"--tag", "echo-v1=latest,echo-v1=current"},
			desiredRevisions: []string{"echo-v2", "echo-v1", "echo-v1"}, // appends two new targets
			desiredTags:      []string{"", "latest", "current"},         // with new tags requested
			desiredPercents:  []int64{100, 0, 0},                        // and assign 0% to each
		},
		{
			name:             "re-assign same tag 'current' to @latest",
			existingTraffic:  append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("current", "", 100, true)),
			inputFlags:       []string{"--tag", "@latest=current"},
			desiredRevisions: []string{""},
			desiredTags:      []string{"current"}, // since no change, no error
			desiredPercents:  []int64{100},
		},
		{
			name:             "assign echo-v1 10% traffic adjusting rest to @latest, echo-v1 isn't present in existing traffic block",
			existingTraffic:  append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("", "", 100, true)),
			inputFlags:       []string{"--traffic", "echo-v1=10,@latest=90"},
			desiredRevisions: []string{"", "echo-v1"},
			desiredTags:      []string{"", ""}, // since no change, no error
			desiredPercents:  []int64{90, 10},
		},
		{
			name:             "traffic split with sum < 100 should work for n-1 revisions",
			existingTraffic:  append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("", "rev-00001", 0, false), newTarget("", "rev-00002", 0, false), newTarget("", "rev-00003", 100, false)),
			inputFlags:       []string{"--traffic", "rev-00001=10,rev-00002=10"},
			desiredRevisions: []string{"rev-00001", "rev-00002", "rev-00003"},
			desiredPercents:  []int64{10, 10, 80},
			desiredTags:      []string{"", "", ""},
			existingRevisions: []servingv1.Revision{{
				ObjectMeta: metav1.ObjectMeta{
					Name: "rev-00001",
					Labels: map[string]string{
						"serving.knative.dev/service": "serviceName",
					},
				},
			}, {
				ObjectMeta: metav1.ObjectMeta{
					Name: "rev-00002",
					Labels: map[string]string{
						"serving.knative.dev/service": "serviceName",
					},
				},
			}, {
				ObjectMeta: metav1.ObjectMeta{
					Name: "rev-00003",
					Labels: map[string]string{
						"serving.knative.dev/service": "serviceName",
					},
				},
			}},
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
			targets, err := Compute(testCmd, testCase.existingTraffic, tFlags, "serviceName", testCase.existingRevisions)
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
			name:            "invalid format for --traffic option",
			existingTraffic: append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("", "", 100, true)),
			inputFlags:      []string{"--traffic", "@latest=100=latest"},
			errMsg:          "expecting the value format in value1=value2, given @latest=100=latest",
		},
		{
			name:            "invalid format for --tag option",
			existingTraffic: append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("", "", 100, true)),
			inputFlags:      []string{"--tag", "@latest="},
			errMsg:          "expecting the value format in value1=value2, given @latest=",
		},
		{
			name:            "repeatedly splitting traffic to @latest revision",
			existingTraffic: append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("", "", 100, true)),
			inputFlags:      []string{"--traffic", "@latest=90,@latest=10"},
			errMsg:          "repetition of identifier @latest is not allowed, use only once with --traffic flag",
		},
		{
			name:            "repeatedly tagging to @latest revision not allowed",
			existingTraffic: append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("", "", 100, true)),
			inputFlags:      []string{"--tag", "@latest=latest,@latest=2"},
			errMsg:          "repetition of identifier @latest is not allowed, use only once with --tag flag",
		},
		{
			name:            "overwriting tag not allowed, to @latest from other revision",
			existingTraffic: append(append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("latest", "", 2, true)), newTarget("stable", "echo-v2", 98, false)),
			inputFlags:      []string{"--tag", "@latest=stable"},
			errMsg:          "refusing to overwrite existing tag in service, add flag '--untag stable' in command to untag it",
		},
		{
			name:            "overwriting tag not allowed, to a revision from other revision",
			existingTraffic: append(append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("latest", "", 2, true)), newTarget("stable", "echo-v2", 98, false)),
			inputFlags:      []string{"--tag", "echo-v2=latest"},
			errMsg:          "refusing to overwrite existing tag in service, add flag '--untag latest' in command to untag it",
		},
		{
			name:            "overwriting tag of @latest not allowed, existing != requested",
			existingTraffic: append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("candidate", "", 100, true)),
			inputFlags:      []string{"--tag", "@latest=current"},
			errMsg:          "tag 'candidate' exists on latest ready revision of service, refusing to overwrite existing tag with 'current', add flag '--untag candidate' in command to untag it",
		},
		{
			name:            "verify error for non integer values given to percent",
			existingTraffic: append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("", "", 100, true)),
			inputFlags:      []string{"--traffic", "@latest=100p"},
			errMsg:          "error converting given 100p to integer value for traffic distribution",
		},
		{
			name:            "verify error for traffic sum not equal to 100",
			existingTraffic: append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("", "", 100, true)),
			inputFlags:      []string{"--traffic", "@latest=40,echo-v1=70"},
			errMsg:          "given traffic percents sum to 110, want 100",
		},
		{
			name:            "verify error for values out of range given to percent",
			existingTraffic: append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("", "", 100, true)),
			inputFlags:      []string{"--traffic", "@latest=-100"},
			errMsg:          "invalid value for traffic percent -100, expected 0 <= percent <= 100",
		},
		{
			name:            "repeatedly splitting traffic to the same revision",
			existingTraffic: append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("", "", 100, true)),
			inputFlags:      []string{"--traffic", "echo-v1=40", "--traffic", "echo-v1=60"},
			errMsg:          "repetition of revision reference echo-v1 is not allowed, use only once with --traffic flag",
		},
		{
			name:            "untag single tag that does not exist",
			existingTraffic: append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("latest", "echo-v1", 100, false)),
			inputFlags:      []string{"--untag", "foo"},
			errMsg:          "tag(s) foo not present for any revisions of service serviceName",
		},
		{
			name:            "untag multiple tags that do not exist",
			existingTraffic: append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("latest", "echo-v1", 100, false)),
			inputFlags:      []string{"--untag", "foo", "--untag", "bar"},
			errMsg:          "tag(s) foo, bar not present for any revisions of service serviceName",
		},
		{
			name:            "traffic split sum < 100 should have N-1 revisions specified",
			existingTraffic: append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("", "rev-00001", 0, false), newTarget("", "rev-00002", 0, false), newTarget("", "rev-00003", 100, false)),
			inputFlags:      []string{"--traffic", "rev-00001=10"},
			errMsg:          errorTrafficDistribution(10, errorDistributionRevisionCount).Error(),
			existingRevisions: []servingv1.Revision{{
				ObjectMeta: metav1.ObjectMeta{
					Name: "rev-00001",
					Labels: map[string]string{
						"serving.knative.dev/service": "serviceName",
					},
				},
			}, {
				ObjectMeta: metav1.ObjectMeta{
					Name: "rev-00002",
					Labels: map[string]string{
						"serving.knative.dev/service": "serviceName",
					},
				},
			}, {
				ObjectMeta: metav1.ObjectMeta{
					Name: "rev-00003",
					Labels: map[string]string{
						"serving.knative.dev/service": "serviceName",
					},
				},
			}},
		},
		{
			name:            "traffic split sum < 100 should not have @latest specified",
			existingTraffic: append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("", "rev-00001", 0, false), newTarget("", "rev-00002", 0, false), newTarget("", "rev-00003", 100, true)),
			inputFlags:      []string{"--traffic", "rev-00001=10,@latest=20"},
			errMsg:          errorTrafficDistribution(30, errorDistributionLatestTag).Error(),
			existingRevisions: []servingv1.Revision{{
				ObjectMeta: metav1.ObjectMeta{
					Name: "rev-00001",
					Labels: map[string]string{
						"serving.knative.dev/service": "serviceName",
					},
				},
			}, {
				ObjectMeta: metav1.ObjectMeta{
					Name: "rev-00002",
					Labels: map[string]string{
						"serving.knative.dev/service": "serviceName",
					},
				},
			}, {
				ObjectMeta: metav1.ObjectMeta{
					Name: "rev-00003",
					Labels: map[string]string{
						"serving.knative.dev/service": "serviceName",
					},
				},
			}},
		},
		{
			name:            "traffic split sum < 100 error when remaining revision not found",
			existingTraffic: append(newServiceTraffic([]servingv1.TrafficTarget{}), newTarget("rev-00003", "rev-00001", 0, false), newTarget("", "rev-00002", 0, false), newTarget("", "rev-00003", 100, true)),
			inputFlags:      []string{"--traffic", "rev-00003=10,rev-00002=20"},
			errMsg:          errorTrafficDistribution(30, errorDistributionRevisionNotFound).Error(),
			existingRevisions: []servingv1.Revision{{
				ObjectMeta: metav1.ObjectMeta{
					Name: "rev-00001",
					Labels: map[string]string{
						"serving.knative.dev/service": "serviceName",
					},
					Annotations: map[string]string{
						revision.RevisionTagsAnnotation: "rev-00003",
					},
				},
			}, {
				ObjectMeta: metav1.ObjectMeta{
					Name: "rev-00002",
					Labels: map[string]string{
						"serving.knative.dev/service": "serviceName",
					},
				},
			}, {
				ObjectMeta: metav1.ObjectMeta{
					Name: "rev-00003",
					Labels: map[string]string{
						"serving.knative.dev/service": "serviceName",
					},
				},
			}},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			testCmd, tFlags := newTestTrafficCommand()
			testCmd.SetArgs(testCase.inputFlags)
			testCmd.Execute()
			_, err := Compute(testCmd, testCase.existingTraffic, tFlags, "serviceName", testCase.existingRevisions)
			assert.Error(t, err, testCase.errMsg)
		})
	}
}
