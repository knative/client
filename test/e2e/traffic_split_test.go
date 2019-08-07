// Copyright 2019 The Knative Authors

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or im
// See the License for the specific language governing permissions and
// limitations under the License.

// +build e2e

package e2e

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/knative/client/pkg/util"
	"gotest.tools/assert"
)

var targetsSeparator = "|"
var targetFieldsSeparator = ","
var targetFieldsLength = 4

// returns deployed service targets separated by '|' and each target fields seprated by comma
var targetsJsonPath = "jsonpath={range .status.traffic[*]}{.tag}{','}{.revisionName}{','}{.percent}{','}{.latestRevision}{'|'}{end}"

// returns deployed service latest revision name
var latestRevisionJsonPath = "jsonpath={.status.traffic[?(@.latestRevision==true)].revisionName}"

// TargetFileds are used in e2e to store expected fields per traffic target
// and actual traffic targets fields of deployed service are converted into struct before comparing
type TargetFields struct {
	Tag      string
	Revision string
	Percent  int
	Latest   bool
}

func newTargetFields(tag, revision string, percent int, latest bool) TargetFields {
	return TargetFields{tag, revision, percent, latest}
}

func splitTargets(s, separator string, partsCount int) ([]string, error) {
	parts := strings.SplitN(s, separator, partsCount)
	if len(parts) != partsCount {
		return nil, errors.New(fmt.Sprintf("expecting to receive parts of length %d, got %d "+
			"string: %s seprator: %s", partsCount, len(parts), s, separator))
	}
	return parts, nil
}

// formatActualTargets takes the traffic targets string received after jsonpath operation and converts
// them into []TargetFields for comparison
func formatActualTargets(t *testing.T, actualTargets []string) (formattedTargets []TargetFields) {
	for _, each := range actualTargets {
		each := strings.TrimSuffix(each, targetFieldsSeparator)
		fields, err := splitTargets(each, targetFieldsSeparator, targetFieldsLength)
		assert.NilError(t, err)
		percentInt, err := strconv.Atoi(fields[2])
		assert.NilError(t, err)
		latestBool, err := strconv.ParseBool(fields[3])
		assert.NilError(t, err)
		formattedTargets = append(formattedTargets, newTargetFields(fields[0], fields[1], percentInt, latestBool))
	}
	return
}

// TestTrafficSplit runs different e2e tests for service traffic splitting and verifies the traffic targets from service status
func TestTrafficSplit(t *testing.T) {
	t.Parallel()
	test := NewE2eTest(t)
	test.Setup(t)
	defer test.Teardown(t)

	serviceBase := "echo"
	t.Run("tag two revisions as v1 and v2 and give 50-50% share",
		func(t *testing.T) {
			serviceName := getServiceNameAndIncrement(serviceBase)
			test.serviceCreate(t, serviceName)
			test.serviceUpdateWithOptions(t, serviceName, []string{"--env", "TARGET=v1"})
			rev1 := test.latestRevisionOfService(t, serviceName)
			test.serviceUpdateWithOptions(t, serviceName, []string{"--env", "TARGET=v2"})
			rev2 := test.latestRevisionOfService(t, serviceName)

			tflags := []string{"--tag", fmt.Sprintf("%s=v1,%s=v2", rev1, rev2),
				"--traffic", "v1=50,v2=50"}
			test.serviceUpdateWithOptions(t, serviceName, tflags)

			// make ordered fields per tflags (tag, revision, percent, latest)
			expectedTargets := []TargetFields{newTargetFields("v1", rev1, 50, false), newTargetFields("v2", rev2, 50, false)}
			test.verifyTargets(t, serviceName, expectedTargets)
			test.serviceDelete(t, serviceName)
		},
	)
	t.Run("ramp/up down a revision to 20% adjusting other traffic to accommodate",
		func(t *testing.T) {
			serviceName := getServiceNameAndIncrement(serviceBase)
			test.serviceCreate(t, serviceName)
			test.serviceUpdateWithOptions(t, serviceName, []string{"--env", "TARGET=v1"})
			rev1 := test.latestRevisionOfService(t, serviceName)
			test.serviceUpdateWithOptions(t, serviceName, []string{"--env", "TARGET=v2"})
			rev2 := test.latestRevisionOfService(t, serviceName)

			tflags := []string{"--traffic", fmt.Sprintf("%s=20,%s=80", rev1, rev2)} // traffic by revision name
			test.serviceUpdateWithOptions(t, serviceName, tflags)

			expectedTargets := []TargetFields{newTargetFields("", rev1, 20, false), newTargetFields("", rev2, 80, false)}
			test.verifyTargets(t, serviceName, expectedTargets)
			test.serviceDelete(t, serviceName)
		},
	)
	t.Run("tag a revision as candidate, without otherwise changing any traffic split",
		func(t *testing.T) {
			serviceName := getServiceNameAndIncrement(serviceBase)
			test.serviceCreate(t, serviceName)
			rev1 := test.latestRevisionOfService(t, serviceName)
			test.serviceUpdateWithOptions(t, serviceName, []string{"--env", "TARGET=v1"})
			rev2 := test.latestRevisionOfService(t, serviceName)

			tflags := []string{"--tag", fmt.Sprintf("%s=%s", rev1, "candidate")} // no traffic, append new target with tag in traffic block
			test.serviceUpdateWithOptions(t, serviceName, tflags)

			expectedTargets := []TargetFields{newTargetFields("", rev2, 100, true), newTargetFields("candidate", rev1, 0, false)}
			test.verifyTargets(t, serviceName, expectedTargets)
			test.serviceDelete(t, serviceName)
		},
	)
	t.Run("tag a revision as candidate, set 2% traffic adjusting other traffic to accommodate",
		func(t *testing.T) {
			serviceName := getServiceNameAndIncrement(serviceBase)
			test.serviceCreate(t, serviceName)
			rev1 := test.latestRevisionOfService(t, serviceName)
			test.serviceUpdateWithOptions(t, serviceName, []string{"--env", "TARGET=v1"})
			rev2 := test.latestRevisionOfService(t, serviceName)

			tflags := []string{"--tag", fmt.Sprintf("%s=%s", rev1, "candidate"),
				"--traffic", "candidate=2%,@latest=98%"} // traffic by tag name and use % at the end
			test.serviceUpdateWithOptions(t, serviceName, tflags)

			expectedTargets := []TargetFields{newTargetFields("", rev2, 98, true), newTargetFields("candidate", rev1, 2, false)}
			test.verifyTargets(t, serviceName, expectedTargets)
			test.serviceDelete(t, serviceName)
		},
	)
	t.Run("update tag for a revision from candidate to current, tag current is present on another revision",
		func(t *testing.T) {
			serviceName := getServiceNameAndIncrement(serviceBase)
			// make available 3 revisions for service first
			test.serviceCreate(t, serviceName)
			rev1 := test.latestRevisionOfService(t, serviceName)

			test.serviceUpdateWithOptions(t, serviceName, []string{"--env", "TARGET=v2"})
			rev2 := test.latestRevisionOfService(t, serviceName)

			test.serviceUpdateWithOptions(t, serviceName, []string{"--env", "TARGET=v3"}) //note that this gives 100% traffic to latest revision (rev3)
			rev3 := test.latestRevisionOfService(t, serviceName)

			// make existing state: tag current and candidate exist in traffic block
			tflags := []string{"--tag", fmt.Sprintf("%s=current,%s=candidate", rev1, rev2)}
			test.serviceUpdateWithOptions(t, serviceName, tflags)

			// desired state of tags: update tag of revision (rev2) from candidate to current (which is present on rev1)
			tflags = []string{"--untag", "current,candidate", "--tag", fmt.Sprintf("%s=current", rev2)} //untag first to update
			test.serviceUpdateWithOptions(t, serviceName, tflags)

			// there will be 2 targets in existing block 1. @latest, 2.for revision $rev2
			// target for rev1 is removed as it had no traffic and we untagged it's tag current
			expectedTargets := []TargetFields{newTargetFields("", rev3, 100, true), newTargetFields("current", rev2, 0, false)}
			test.verifyTargets(t, serviceName, expectedTargets)
			test.serviceDelete(t, serviceName)
		},
	)
	t.Run("update tag from testing to staging for @latest revision",
		func(t *testing.T) {
			serviceName := getServiceNameAndIncrement(serviceBase)
			test.serviceCreate(t, serviceName)
			rev1 := test.latestRevisionOfService(t, serviceName)

			// make existing state: tag @latest as testing
			tflags := []string{"--tag", "@latest=testing"}
			test.serviceUpdateWithOptions(t, serviceName, tflags)

			// desired state: change tag from testing to staging
			tflags = []string{"--untag", "testing", "--tag", "@latest=staging"}
			test.serviceUpdateWithOptions(t, serviceName, tflags)

			expectedTargets := []TargetFields{newTargetFields("staging", rev1, 100, true)}
			test.verifyTargets(t, serviceName, expectedTargets)
			test.serviceDelete(t, serviceName)
		},
	)
	t.Run("update tag from testing to staging for a revision (non @latest)",
		func(t *testing.T) {
			serviceName := getServiceNameAndIncrement(serviceBase)
			test.serviceCreate(t, serviceName)
			rev1 := test.latestRevisionOfService(t, serviceName)

			test.serviceUpdateWithOptions(t, serviceName, []string{"--env", "TARGET=v2"})
			rev2 := test.latestRevisionOfService(t, serviceName)

			// make existing state: tag a revision as testing
			tflags := []string{"--tag", fmt.Sprintf("%s=testing", rev1)}
			test.serviceUpdateWithOptions(t, serviceName, tflags)

			// desired state: change tag from testing to staging
			tflags = []string{"--untag", "testing", "--tag", fmt.Sprintf("%s=staging", rev1)}
			test.serviceUpdateWithOptions(t, serviceName, tflags)

			expectedTargets := []TargetFields{newTargetFields("", rev2, 100, true),
				newTargetFields("staging", rev1, 0, false)}
			test.verifyTargets(t, serviceName, expectedTargets)
			test.serviceDelete(t, serviceName)
		},
	)
	// test reducing number of targets from traffic blockdd
	t.Run("remove a revision with tag old from traffic block entierly",
		func(t *testing.T) {
			serviceName := getServiceNameAndIncrement(serviceBase)
			test.serviceCreate(t, serviceName)
			rev1 := test.latestRevisionOfService(t, serviceName)

			test.serviceUpdateWithOptions(t, serviceName, []string{"--env", "TARGET=v2"})
			rev2 := test.latestRevisionOfService(t, serviceName)

			// existing state: traffic block having a revision with tag old and some traffic
			tflags := []string{"--tag", fmt.Sprintf("%s=old", rev1),
				"--traffic", "old=2,@latest=98"}
			test.serviceUpdateWithOptions(t, serviceName, tflags)

			// desired state: remove revision with tag old
			tflags = []string{"--untag", "old", "--traffic", "@latest=100"}
			test.serviceUpdateWithOptions(t, serviceName, tflags)

			expectedTargets := []TargetFields{newTargetFields("", rev2, 100, true)}
			test.verifyTargets(t, serviceName, expectedTargets)
			test.serviceDelete(t, serviceName)
		},
	)
	t.Run("tag a revision as stable and current with 50-50% traffic",
		func(t *testing.T) {
			serviceName := getServiceNameAndIncrement(serviceBase)
			test.serviceCreate(t, serviceName)
			rev1 := test.latestRevisionOfService(t, serviceName)

			// existing state: traffic block having two targets
			test.serviceUpdateWithOptions(t, serviceName, []string{"--env", "TARGET=v2"})

			// desired state: tag non-@latest revision with two tags and 50-50% traffic each
			tflags := []string{"--tag", fmt.Sprintf("%s=stable,%s=current", rev1, rev1),
				"--traffic", "stable=50%,current=50%"}
			test.serviceUpdateWithOptions(t, serviceName, tflags)

			expectedTargets := []TargetFields{newTargetFields("stable", rev1, 50, false), newTargetFields("current", rev1, 50, false)}
			test.verifyTargets(t, serviceName, expectedTargets)
			test.serviceDelete(t, serviceName)
		},
	)
	t.Run("revert all traffic to latest ready revision of service",
		func(t *testing.T) {
			serviceName := getServiceNameAndIncrement(serviceBase)
			test.serviceCreate(t, serviceName)
			rev1 := test.latestRevisionOfService(t, serviceName)

			test.serviceUpdateWithOptions(t, serviceName, []string{"--env", "TARGET=v2"})
			rev2 := test.latestRevisionOfService(t, serviceName)

			// existing state: latest revision not getting any traffic
			tflags := []string{"--traffic", fmt.Sprintf("%s=100", rev1)}
			test.serviceUpdateWithOptions(t, serviceName, tflags)

			// desired state: revert traffic to latest revision
			tflags = []string{"--traffic", "@latest=100"}
			test.serviceUpdateWithOptions(t, serviceName, tflags)

			expectedTargets := []TargetFields{newTargetFields("", rev2, 100, true)}
			test.verifyTargets(t, serviceName, expectedTargets)
			test.serviceDelete(t, serviceName)
		},
	)
	t.Run("tag latest ready revision of service as current",
		func(t *testing.T) {
			serviceName := getServiceNameAndIncrement(serviceBase)
			// existing state: latest revision has no tag
			test.serviceCreate(t, serviceName)
			rev1 := test.latestRevisionOfService(t, serviceName)

			// desired state: tag current to latest ready revision
			tflags := []string{"--tag", "@latest=current"}
			test.serviceUpdateWithOptions(t, serviceName, tflags)

			expectedTargets := []TargetFields{newTargetFields("current", rev1, 100, true)}
			test.verifyTargets(t, serviceName, expectedTargets)
			test.serviceDelete(t, serviceName)
		},
	)
	t.Run("update tag for a revision as testing and assign all the traffic to it:",
		func(t *testing.T) {
			serviceName := getServiceNameAndIncrement(serviceBase)
			test.serviceCreate(t, serviceName)
			rev1 := test.latestRevisionOfService(t, serviceName)

			test.serviceUpdateWithOptions(t, serviceName, []string{"--env", "TARGET=v2"})
			rev2 := test.latestRevisionOfService(t, serviceName)

			// existing state: two revision exists with traffic share and
			// each revision has tag and traffic portions
			tflags := []string{"--tag", fmt.Sprintf("@latest=current,%s=candidate", rev1),
				"--traffic", "current=90,candidate=10"}
			test.serviceUpdateWithOptions(t, serviceName, tflags)

			// desired state: update tag for rev1 as testing (from candidate) with 100% traffic
			tflags = []string{"--untag", "candidate", "--tag", fmt.Sprintf("%s=testing", rev1),
				"--traffic", "testing=100"}
			test.serviceUpdateWithOptions(t, serviceName, tflags)

			expectedTargets := []TargetFields{newTargetFields("current", rev2, 0, true),
				newTargetFields("testing", rev1, 100, false)}
			test.verifyTargets(t, serviceName, expectedTargets)
			test.serviceDelete(t, serviceName)
		},
	)
	t.Run("replace latest tag of a revision with old and give latest to another revision",
		func(t *testing.T) {
			serviceName := getServiceNameAndIncrement(serviceBase)
			test.serviceCreate(t, serviceName)
			rev1 := test.latestRevisionOfService(t, serviceName)

			test.serviceUpdateWithOptions(t, serviceName, []string{"--env", "TARGET=v2"})
			rev2 := test.latestRevisionOfService(t, serviceName)

			// existing state: a revision exist with latest tag
			tflags := []string{"--tag", fmt.Sprintf("%s=latest", rev1)}
			test.serviceUpdateWithOptions(t, serviceName, tflags)

			// desired state of revision tags: rev1=old rev2=latest
			tflags = []string{"--untag", "latest", "--tag", fmt.Sprintf("%s=old,%s=latest", rev1, rev2)}
			test.serviceUpdateWithOptions(t, serviceName, tflags)

			expectedTargets := []TargetFields{newTargetFields("", rev2, 100, true),
				newTargetFields("old", rev1, 0, false),
				// Tagging by revision name adds a new target even though latestReadyRevision==rev2,
				// because we didn't refer @latest reference, but explcit name of revision.
				// In spec of traffic block (not status) either latestReadyRevision:true or revisionName can be given per target
				newTargetFields("latest", rev2, 0, false)}

			test.verifyTargets(t, serviceName, expectedTargets)
			test.serviceDelete(t, serviceName)
		},
	)
}

func (test *e2eTest) verifyTargets(t *testing.T, serviceName string, expectedTargets []TargetFields) {
	out := test.serviceDescribeWithJsonPath(t, serviceName, targetsJsonPath)
	assert.Check(t, out != "")
	out = strings.TrimSuffix(out, targetsSeparator)
	actualTargets, err := splitTargets(out, targetsSeparator, len(expectedTargets))
	assert.NilError(t, err)
	formattedActualTargets := formatActualTargets(t, actualTargets)
	assert.DeepEqual(t, expectedTargets, formattedActualTargets)
}

func (test *e2eTest) latestRevisionOfService(t *testing.T, serviceName string) string {
	return test.serviceDescribeWithJsonPath(t, serviceName, latestRevisionJsonPath)
}

func (test *e2eTest) serviceDescribeWithJsonPath(t *testing.T, serviceName, jsonpath string) string {
	command := []string{"service", "describe", serviceName, "-o", jsonpath}
	out, err := test.kn.RunWithOpts(command, runOpts{})
	assert.NilError(t, err)
	return out
}

func (test *e2eTest) serviceUpdateWithOptions(t *testing.T, serviceName string, options []string) {
	command := []string{"service", "update", serviceName}
	command = append(command, options...)
	out, err := test.kn.RunWithOpts(command, runOpts{NoNamespace: false})
	assert.NilError(t, err)
	assert.Check(t, util.ContainsAll(out, "Service", serviceName, "update", "namespace", test.kn.namespace))
}
