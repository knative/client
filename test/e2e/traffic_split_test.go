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
// +build !eventing

package e2e

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"gotest.tools/assert"

	"knative.dev/client/lib/test"
)

var targetsSeparator = "|"
var targetFieldsSeparator = ","
var targetFieldsLength = 4

// returns deployed service targets separated by '|' and each target fields separated by comma
var targetsJsonPath = "jsonpath={range .status.traffic[*]}{.tag}{','}{.revisionName}{','}{.percent}{','}{.latestRevision}{'|'}{end}"

// TargetFields are used in e2e to store expected fields per traffic target
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
	s = strings.TrimSuffix(s, targetsSeparator)
	parts := strings.Split(s, separator)
	if len(parts) != partsCount {
		return nil, fmt.Errorf("expecting %d targets, got %d targets "+
			"targets: %s separator: %s", partsCount, len(parts), s, separator)
	}
	return parts, nil
}

// formatActualTargets takes the traffic targets string received after jsonpath operation and converts
// them into []TargetFields for comparison
func formatActualTargets(t *testing.T, it *test.KnTest, actualTargets []string) (formattedTargets []TargetFields) {
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

// TestTrafficSplitSuite runs different e2e tests for service traffic splitting and verifies the traffic targets from service status
func TestTrafficSplit(t *testing.T) {
	t.Parallel()
	it, err := test.NewKnTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, it.Teardown())
	}()

	serviceBase := "echo"
	t.Run("50:50",
		func(t *testing.T) {
			t.Log("tag two revisions as v1 and v2 and give 50-50% share")
			r := test.NewKnRunResultCollector(t, it)
			defer r.DumpIfFailed()

			serviceName := test.GetNextServiceName(serviceBase)
			test.ServiceCreate(r, serviceName)

			rev1 := fmt.Sprintf("%s-rev-1", serviceName)
			test.ServiceUpdate(r, serviceName, "--env", "TARGET=v1", "--revision-name", rev1)

			rev2 := fmt.Sprintf("%s-rev-2", serviceName)
			test.ServiceUpdate(r, serviceName, "--env", "TARGET=v2", "--revision-name", rev2)

			tflags := []string{"--tag", fmt.Sprintf("%s=v1,%s=v2", rev1, rev2),
				"--traffic", "v1=50,v2=50"}
			test.ServiceUpdate(r, serviceName, tflags...)

			// make ordered fields per tflags (tag, revision, percent, latest)
			expectedTargets := []TargetFields{newTargetFields("v1", rev1, 50, false), newTargetFields("v2", rev2, 50, false)}
			verifyTargets(r, serviceName, expectedTargets, false)
			test.ServiceDelete(r, serviceName)
		},
	)
	t.Run("20:80",
		func(t *testing.T) {
			t.Log("ramp/up down a revision to 20% adjusting other traffic to accommodate")
			r := test.NewKnRunResultCollector(t, it)
			defer r.DumpIfFailed()

			serviceName := test.GetNextServiceName(serviceBase)
			test.ServiceCreate(r, serviceName)

			rev1 := fmt.Sprintf("%s-rev-1", serviceName)
			test.ServiceUpdate(r, serviceName, "--env", "TARGET=v1", "--revision-name", rev1)

			rev2 := fmt.Sprintf("%s-rev-2", serviceName)
			test.ServiceUpdate(r, serviceName, "--env", "TARGET=v2", "--revision-name", rev2)

			test.ServiceUpdate(r, serviceName, "--traffic", fmt.Sprintf("%s=20,%s=80", rev1, rev2))

			expectedTargets := []TargetFields{newTargetFields("", rev1, 20, false), newTargetFields("", rev2, 80, false)}
			verifyTargets(r, serviceName, expectedTargets, false)
			test.ServiceDelete(r, serviceName)
		},
	)
	t.Run("TagCandidate",
		func(t *testing.T) {
			t.Log("tag a revision as candidate, without otherwise changing any traffic split")
			r := test.NewKnRunResultCollector(t, it)
			defer r.DumpIfFailed()

			serviceName := test.GetNextServiceName(serviceBase)
			rev1 := fmt.Sprintf("%s-rev-1", serviceName)
			serviceCreateWithOptions(r, serviceName, "--revision-name", rev1)

			rev2 := fmt.Sprintf("%s-rev-2", serviceName)
			test.ServiceUpdate(r, serviceName, "--env", "TARGET=v1", "--revision-name", rev2)

			// no traffic, append new target with tag in traffic block
			test.ServiceUpdate(r, serviceName, "--tag", fmt.Sprintf("%s=%s", rev1, "candidate"))

			expectedTargets := []TargetFields{newTargetFields("", rev2, 100, true), newTargetFields("candidate", rev1, 0, false)}
			verifyTargets(r, serviceName, expectedTargets, false)
			test.ServiceDelete(r, serviceName)
		},
	)
	t.Run("TagCandidate:2:98",
		func(t *testing.T) {
			t.Log("tag a revision as candidate, set 2% traffic adjusting other traffic to accommodate")
			r := test.NewKnRunResultCollector(t, it)
			defer r.DumpIfFailed()

			serviceName := test.GetNextServiceName(serviceBase)
			rev1 := fmt.Sprintf("%s-rev-1", serviceName)
			serviceCreateWithOptions(r, serviceName, "--revision-name", rev1)

			rev2 := fmt.Sprintf("%s-rev-2", serviceName)
			test.ServiceUpdate(r, serviceName, "--env", "TARGET=v1", "--revision-name", rev2)

			// traffic by tag name and use % at the end
			test.ServiceUpdate(r, serviceName,
				"--tag", fmt.Sprintf("%s=%s", rev1, "candidate"),
				"--traffic", "candidate=2%,@latest=98%")

			expectedTargets := []TargetFields{newTargetFields("", rev2, 98, true), newTargetFields("candidate", rev1, 2, false)}
			verifyTargets(r, serviceName, expectedTargets, false)
			test.ServiceDelete(r, serviceName)
		},
	)
	t.Run("TagCurrent",
		func(t *testing.T) {
			t.Log("update tag for a revision from candidate to current, tag current is present on another revision")
			r := test.NewKnRunResultCollector(t, it)
			defer r.DumpIfFailed()

			serviceName := test.GetNextServiceName(serviceBase)
			// make available 3 revisions for service first
			rev1 := fmt.Sprintf("%s-rev-1", serviceName)
			serviceCreateWithOptions(r, serviceName, "--revision-name", rev1)

			rev2 := fmt.Sprintf("%s-rev-2", serviceName)
			test.ServiceUpdate(r, serviceName, "--env", "TARGET=v2", "--revision-name", rev2)

			rev3 := fmt.Sprintf("%s-rev-3", serviceName)
			test.ServiceUpdate(r, serviceName, "--env", "TARGET=v3", "--revision-name", rev3) //note that this gives 100% traffic to latest revision (rev3)

			// make existing state: tag current and candidate exist in traffic block
			test.ServiceUpdate(r, serviceName, "--tag", fmt.Sprintf("%s=current,%s=candidate", rev1, rev2))

			// desired state of tags: update tag of revision (rev2) from candidate to current (which is present on rev1)
			//untag first to update
			test.ServiceUpdate(r, serviceName,
				"--untag", "current,candidate",
				"--tag", fmt.Sprintf("%s=current", rev2))

			// there will be 2 targets in existing block 1. @latest, 2.for revision $rev2
			// target for rev1 is removed as it had no traffic and we untagged it's tag current
			expectedTargets := []TargetFields{newTargetFields("", rev3, 100, true), newTargetFields("current", rev2, 0, false)}
			verifyTargets(r, serviceName, expectedTargets, false)
			test.ServiceDelete(r, serviceName)
		},
	)
	t.Run("TagStagingLatest",
		func(t *testing.T) {
			t.Log("update tag from testing to staging for @latest revision")
			r := test.NewKnRunResultCollector(t, it)
			defer r.DumpIfFailed()

			serviceName := test.GetNextServiceName(serviceBase)
			rev1 := fmt.Sprintf("%s-rev-1", serviceName)
			serviceCreateWithOptions(r, serviceName, "--revision-name", rev1)

			// make existing state: tag @latest as testing
			test.ServiceUpdate(r, serviceName, "--tag", "@latest=testing")

			// desired state: change tag from testing to staging
			test.ServiceUpdate(r, serviceName, "--untag", "testing", "--tag", "@latest=staging")

			expectedTargets := []TargetFields{newTargetFields("staging", rev1, 100, true)}
			verifyTargets(r, serviceName, expectedTargets, false)
			test.ServiceDelete(r, serviceName)
		},
	)
	t.Run("TagStagingNonLatest",
		func(t *testing.T) {
			t.Log("update tag from testing to staging for a revision (non @latest)")
			r := test.NewKnRunResultCollector(t, it)
			defer r.DumpIfFailed()

			serviceName := test.GetNextServiceName(serviceBase)
			rev1 := fmt.Sprintf("%s-rev-1", serviceName)
			serviceCreateWithOptions(r, serviceName, "--revision-name", rev1)

			rev2 := fmt.Sprintf("%s-rev-2", serviceName)
			test.ServiceUpdate(r, serviceName, "--env", "TARGET=v2", "--revision-name", rev2)

			// make existing state: tag a revision as testing
			test.ServiceUpdate(r, serviceName, "--tag", fmt.Sprintf("%s=testing", rev1))

			// desired state: change tag from testing to staging
			test.ServiceUpdate(r, serviceName, "--untag", "testing", "--tag", fmt.Sprintf("%s=staging", rev1))

			expectedTargets := []TargetFields{newTargetFields("", rev2, 100, true),
				newTargetFields("staging", rev1, 0, false)}
			verifyTargets(r, serviceName, expectedTargets, false)
			test.ServiceDelete(r, serviceName)
		},
	)
	// test reducing number of targets from traffic blocked
	t.Run("RemoveTag",
		func(t *testing.T) {
			t.Log("remove a revision with tag old from traffic block entirely")
			r := test.NewKnRunResultCollector(t, it)
			defer r.DumpIfFailed()

			serviceName := test.GetNextServiceName(serviceBase)
			rev1 := fmt.Sprintf("%s-rev-1", serviceName)
			serviceCreateWithOptions(r, serviceName, "--revision-name", rev1)

			rev2 := fmt.Sprintf("%s-rev-2", serviceName)
			test.ServiceUpdate(r, serviceName, "--env", "TARGET=v2", "--revision-name", rev2)

			// existing state: traffic block having a revision with tag old and some traffic
			test.ServiceUpdate(r, serviceName,
				"--tag", fmt.Sprintf("%s=old", rev1),
				"--traffic", "old=2,@latest=98")

			// desired state: remove revision with tag old
			test.ServiceUpdate(r, serviceName, "--untag", "old", "--traffic", "@latest=100")

			expectedTargets := []TargetFields{newTargetFields("", rev2, 100, true)}
			verifyTargets(r, serviceName, expectedTargets, false)
			test.ServiceDelete(r, serviceName)
		},
	)
	t.Run("TagStable:50:50",
		func(t *testing.T) {
			t.Log("tag a revision as stable and current with 50-50% traffic")
			r := test.NewKnRunResultCollector(t, it)
			defer r.DumpIfFailed()

			serviceName := test.GetNextServiceName(serviceBase)
			rev1 := fmt.Sprintf("%s-rev-1", serviceName)
			serviceCreateWithOptions(r, serviceName, "--revision-name", rev1)

			// existing state: traffic block having two targets
			test.ServiceUpdate(r, serviceName, "--env", "TARGET=v2")

			// desired state: tag non-@latest revision with two tags and 50-50% traffic each
			test.ServiceUpdate(r, serviceName,
				"--tag", fmt.Sprintf("%s=stable,%s=current", rev1, rev1),
				"--traffic", "stable=50%,current=50%")

			expectedTargets := []TargetFields{newTargetFields("stable", rev1, 50, false), newTargetFields("current", rev1, 50, false)}
			verifyTargets(r, serviceName, expectedTargets, false)
			test.ServiceDelete(r, serviceName)
		},
	)
	t.Run("RevertToLatest",
		func(t *testing.T) {
			t.Log("revert all traffic to latest ready revision of service")
			r := test.NewKnRunResultCollector(t, it)
			defer r.DumpIfFailed()

			serviceName := test.GetNextServiceName(serviceBase)
			rev1 := fmt.Sprintf("%s-rev-1", serviceName)
			serviceCreateWithOptions(r, serviceName, "--revision-name", rev1)

			rev2 := fmt.Sprintf("%s-rev-2", serviceName)
			test.ServiceUpdate(r, serviceName, "--env", "TARGET=v2", "--revision-name", rev2)

			// existing state: latest ready revision not getting any traffic
			test.ServiceUpdate(r, serviceName, "--traffic", fmt.Sprintf("%s=100", rev1))

			// desired state: revert traffic to latest ready revision
			test.ServiceUpdate(r, serviceName, "--traffic", "@latest=100")

			expectedTargets := []TargetFields{newTargetFields("", rev2, 100, true)}
			verifyTargets(r, serviceName, expectedTargets, false)
			test.ServiceDelete(r, serviceName)
		},
	)
	t.Run("TagLatestAsCurrent",
		func(t *testing.T) {
			t.Log("tag latest ready revision of service as current")
			r := test.NewKnRunResultCollector(t, it)
			defer r.DumpIfFailed()

			serviceName := test.GetNextServiceName(serviceBase)
			// existing state: latest revision has no tag
			rev1 := fmt.Sprintf("%s-rev-1", serviceName)
			serviceCreateWithOptions(r, serviceName, "--revision-name", rev1)

			// desired state: tag latest ready revision as 'current'
			test.ServiceUpdate(r, serviceName, "--tag", "@latest=current")

			expectedTargets := []TargetFields{newTargetFields("current", rev1, 100, true)}
			verifyTargets(r, serviceName, expectedTargets, false)
			test.ServiceDelete(r, serviceName)
		},
	)
	t.Run("UpdateTag:100:0",
		func(t *testing.T) {
			t.Log("update tag for a revision as testing and assign all the traffic to it")
			r := test.NewKnRunResultCollector(t, it)
			defer r.DumpIfFailed()

			serviceName := test.GetNextServiceName(serviceBase)
			rev1 := fmt.Sprintf("%s-rev-1", serviceName)
			serviceCreateWithOptions(r, serviceName, "--revision-name", rev1)

			rev2 := fmt.Sprintf("%s-rev-2", serviceName)
			test.ServiceUpdate(r, serviceName, "--env", "TARGET=v2", "--revision-name", rev2)

			// existing state: two revision exists with traffic share and
			// each revision has tag and traffic portions
			test.ServiceUpdate(r, serviceName,
				"--tag", fmt.Sprintf("@latest=current,%s=candidate", rev1),
				"--traffic", "current=90,candidate=10")

			// desired state: update tag for rev1 as testing (from candidate) with 100% traffic
			test.ServiceUpdate(r, serviceName,
				"--untag", "candidate", "--tag", fmt.Sprintf("%s=testing", rev1),
				"--traffic", "testing=100")

			expectedTargets := []TargetFields{newTargetFields("current", rev2, 0, true),
				newTargetFields("testing", rev1, 100, false)}
			verifyTargets(r, serviceName, expectedTargets, false)
			test.ServiceDelete(r, serviceName)
		},
	)
	t.Run("TagReplace",
		func(t *testing.T) {
			t.Log("replace latest tag of a revision with old and give latest to another revision")
			r := test.NewKnRunResultCollector(t, it)
			defer r.DumpIfFailed()

			serviceName := test.GetNextServiceName(serviceBase)
			rev1 := fmt.Sprintf("%s-rev-1", serviceName)
			serviceCreateWithOptions(r, serviceName, "--revision-name", rev1)

			rev2 := fmt.Sprintf("%s-rev-2", serviceName)
			test.ServiceUpdate(r, serviceName, "--env", "TARGET=v2", "--revision-name", rev2)

			// existing state: a revision exist with latest tag
			test.ServiceUpdate(r, serviceName, "--tag", fmt.Sprintf("%s=latest", rev1))

			// desired state of revision tags: rev1=old rev2=latest
			test.ServiceUpdate(r, serviceName,
				"--untag", "latest",
				"--tag", fmt.Sprintf("%s=old,%s=latest", rev1, rev2))

			expectedTargets := []TargetFields{newTargetFields("", rev2, 100, true),
				newTargetFields("old", rev1, 0, false),
				// Tagging by revision name adds a new target even though latestReadyRevision==rev2,
				// because we didn't refer @latest reference, but explicit name of revision.
				// In spec of traffic block (not status) either latestReadyRevision:true or revisionName can be given per target
				newTargetFields("latest", rev2, 0, false)}

			verifyTargets(r, serviceName, expectedTargets, false)
			test.ServiceDelete(r, serviceName)
		},
	)
	t.Run("UntagNonExistentTag",
		func(t *testing.T) {
			t.Log("use --untag on a tag that does not exist")
			r := test.NewKnRunResultCollector(t, it)
			defer r.DumpIfFailed()

			serviceName := test.GetNextServiceName(serviceBase)
			rev1 := fmt.Sprintf("%s-rev-1", serviceName)
			serviceCreateWithOptions(r, serviceName, "--revision-name", rev1)

			// existing state: a revision exist with latest tag
			test.ServiceUpdate(r, serviceName, "--tag", fmt.Sprintf("%s=latest", rev1))

			// desired state of revision tags: rev1=latest
			// attempt to untag a tag that does not exist for the service's revisions
			test.ServiceUpdateWithError(r, serviceName, "--untag", "foo")

			// state should remain the same as error from --untag will stop service update
			expectedTargets := []TargetFields{
				newTargetFields("", rev1, 100, true),
				newTargetFields("latest", rev1, 0, false),
			}

			verifyTargets(r, serviceName, expectedTargets, true)
			test.ServiceDelete(r, serviceName)
		},
	)
	t.Run("UntagNonExistentTagAndValidTag",
		func(t *testing.T) {
			t.Log("use --untag on a tag that does not exist in addition to a tag that exists")
			r := test.NewKnRunResultCollector(t, it)
			defer r.DumpIfFailed()

			serviceName := test.GetNextServiceName(serviceBase)
			rev1 := fmt.Sprintf("%s-rev-1", serviceName)
			serviceCreateWithOptions(r, serviceName, "--revision-name", rev1)

			// existing state: a revision exist with latest tag
			test.ServiceUpdate(r, serviceName, "--tag", fmt.Sprintf("%s=latest", rev1))

			// desired state of revision tags: rev1=latest
			// attempt to untag a tag that does not exist for the service's revisions (foo)
			// and also untag a tag that exists (latest)
			test.ServiceUpdateWithError(r, serviceName, "--untag", "latest", "--untag", "foo")

			// state should remain the same as error from --untag will stop service update
			expectedTargets := []TargetFields{
				newTargetFields("", rev1, 100, true),
				newTargetFields("latest", rev1, 0, false),
			}

			verifyTargets(r, serviceName, expectedTargets, true)
			test.ServiceDelete(r, serviceName)
		},
	)
}

func verifyTargets(r *test.KnRunResultCollector, serviceName string, expectedTargets []TargetFields, expectErr bool) {
	out := test.ServiceDescribeWithJSONPath(r, serviceName, targetsJsonPath)
	assert.Check(r.T(), out != "")
	actualTargets, err := splitTargets(out, targetsSeparator, len(expectedTargets))
	if !expectErr {
		assert.NilError(r.T(), err)
	}
	formattedActualTargets := formatActualTargets(r.T(), r.KnTest(), actualTargets)
	assert.DeepEqual(r.T(), expectedTargets, formattedActualTargets)
	if r.T().Failed() {
		r.AddDump("service", serviceName, r.KnTest().Kn().Namespace())
	}
}
