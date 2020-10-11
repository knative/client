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
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"knative.dev/pkg/ptr"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"knative.dev/client/pkg/kn/commands/flags"
)

var latestRevisionRef = "@latest"

// ServiceTraffic type for operating on service traffic targets
type ServiceTraffic []servingv1.TrafficTarget

func newServiceTraffic(traffic []servingv1.TrafficTarget) ServiceTraffic {
	return ServiceTraffic(traffic)
}

func splitByEqualSign(pair string) (string, string, error) {
	parts := strings.Split(pair, "=")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("expecting the value format in value1=value2, given %s", pair)
	}
	return parts[0], strings.TrimSuffix(parts[1], "%"), nil
}

func newTarget(tag, revision string, percent int64, latestRevision bool) (target servingv1.TrafficTarget) {
	target.Percent = ptr.Int64(percent)
	target.Tag = tag
	if latestRevision {
		target.LatestRevision = ptr.Bool(true)
	} else {
		// as LatestRevision and RevisionName can't be specified together for a target
		target.LatestRevision = ptr.Bool(false)
		target.RevisionName = revision
	}
	return
}

func (e ServiceTraffic) isTagPresentOnRevision(tag, revision string) bool {
	for _, target := range e {
		if target.Tag == tag && target.RevisionName == revision {
			return true
		}
	}
	return false
}

func (e ServiceTraffic) tagOfLatestReadyRevision() string {
	for _, target := range e {
		if *target.LatestRevision {
			return target.Tag
		}
	}
	return ""
}

func (e ServiceTraffic) isTagPresent(tag string) bool {
	for _, target := range e {
		if target.Tag == tag {
			return true
		}
	}
	return false
}

func (e ServiceTraffic) untagRevision(tag string, serviceName string) bool {
	for i, target := range e {
		if target.Tag == tag {
			e[i].Tag = ""
			return true
		}
	}

	return false
}

func (e ServiceTraffic) isRevisionPresent(revision string) bool {
	for _, target := range e {
		if target.RevisionName == revision {
			return true
		}
	}
	return false
}

func (e ServiceTraffic) isLatestRevisionTrue() bool {
	for _, target := range e {
		if *target.LatestRevision {
			return true
		}
	}
	return false
}

// TagRevision assigns given tag to a revision
func (e ServiceTraffic) TagRevision(tag, revision string) ServiceTraffic {
	for i, target := range e {
		if target.RevisionName == revision {
			if target.Tag != "" { // referenced revision is requested to have multiple tags
				break
			} else {
				e[i].Tag = tag // referenced revision doesn't have tag, tag it
				return e
			}
		}
	}
	// append a new target if revision doesn't exist in traffic block
	// or if referenced revision is requested to have multiple tags
	e = append(e, newTarget(tag, revision, 0, false))
	return e
}

// TagLatestRevision assigns given tag to latest ready revision
func (e ServiceTraffic) TagLatestRevision(tag string) ServiceTraffic {
	for i, target := range e {
		if *target.LatestRevision {
			e[i].Tag = tag
			return e
		}
	}
	e = append(e, newTarget(tag, "", 0, true))
	return e
}

// SetTrafficByRevision checks given revision in existing traffic block and sets given percent if found
func (e ServiceTraffic) SetTrafficByRevision(revision string, percent int64) {
	for i, target := range e {
		if target.RevisionName == revision {
			e[i].Percent = ptr.Int64(percent)
			break
		}
	}
}

// SetTrafficByTag checks given tag in existing traffic block and sets given percent if found
func (e ServiceTraffic) SetTrafficByTag(tag string, percent int64) {
	for i, target := range e {
		if target.Tag == tag {
			e[i].Percent = ptr.Int64(percent)
			break
		}
	}
}

// SetTrafficByLatestRevision sets given percent to latest ready revision of service
func (e ServiceTraffic) SetTrafficByLatestRevision(percent int64) {
	for i, target := range e {
		if *target.LatestRevision {
			e[i].Percent = ptr.Int64(percent)
			break
		}
	}
}

// ResetAllTargetPercent resets (0) 'Percent' field for all the traffic targets
func (e ServiceTraffic) ResetAllTargetPercent() {
	for i := range e {
		e[i].Percent = ptr.Int64(0)
	}
}

// RemoveNullTargets removes targets from traffic block
// if they don't have a tag and 0 percent traffic
func (e ServiceTraffic) RemoveNullTargets() (newTraffic ServiceTraffic) {
	for _, target := range e {
		if target.Tag != "" || *target.Percent != int64(0) {
			newTraffic = append(newTraffic, target)
		}
	}

	return newTraffic
}

func errorOverWritingtagOfLatestReadyRevision(existingTag, requestedTag string) error {
	return fmt.Errorf("tag '%s' exists on latest ready revision of service, "+
		"refusing to overwrite existing tag with '%s', "+
		"add flag '--untag %s' in command to untag it", existingTag, requestedTag, existingTag)
}

func errorOverWritingTag(tag string) error {
	return fmt.Errorf("refusing to overwrite existing tag in service, "+
		"add flag '--untag %s' in command to untag it", tag)
}

func errorRepeatingRevision(forFlag string, name string) error {
	if name == latestRevisionRef {
		name = "identifier " + latestRevisionRef
	} else {
		name = "revision reference " + name
	}
	return fmt.Errorf("repetition of %s "+
		"is not allowed, use only once with %s flag", name, forFlag)
}

// verifies if user has repeated @latest field in --tag or --traffic flags
// verifyInputSanity checks:
// - if user has repeated @latest field in --tag or --traffic flags
// - if provided traffic portion are integers
func verifyInputSanity(trafficFlags *flags.Traffic) error {
	var latestRevisionTag = false
	var sum = 0

	for _, each := range trafficFlags.RevisionsTags {
		revision, _, err := splitByEqualSign(each)
		if err != nil {
			return err
		}

		if latestRevisionTag && revision == latestRevisionRef {
			return errorRepeatingRevision("--tag", latestRevisionRef)
		}

		if revision == latestRevisionRef {
			latestRevisionTag = true
		}
	}

	revisionRefMap := make(map[string]int)
	for i, each := range trafficFlags.RevisionsPercentages {
		revisionRef, percent, err := splitByEqualSign(each)
		if err != nil {
			return err
		}

		// To check if there are duplicate revision names in traffic flags
		if _, exist := revisionRefMap[revisionRef]; exist {
			return errorRepeatingRevision("--traffic", revisionRef)
		}
		revisionRefMap[revisionRef] = i

		percentInt, err := strconv.Atoi(percent)
		if err != nil {
			return fmt.Errorf("error converting given %s to integer value for traffic distribution", percent)
		}

		if percentInt < 0 || percentInt > 100 {
			return fmt.Errorf("invalid value for traffic percent %d, expected 0 <= percent <= 100", percentInt)
		}

		sum += percentInt
	}

	// equivalent check for `cmd.Flags().Changed("traffic")` as we don't have `cmd` in this function
	if len(trafficFlags.RevisionsPercentages) > 0 && sum != 100 {
		return fmt.Errorf("given traffic percents sum to %d, want 100", sum)
	}

	return nil
}

// Compute takes service traffic targets and updates per given traffic flags
func Compute(cmd *cobra.Command, targets []servingv1.TrafficTarget,
	trafficFlags *flags.Traffic, serviceName string) ([]servingv1.TrafficTarget, error) {
	err := verifyInputSanity(trafficFlags)
	if err != nil {
		return nil, err
	}

	traffic := newServiceTraffic(targets)

	// First precedence: Untag revisions
	var errTagNames []string
	for _, tag := range trafficFlags.UntagRevisions {
		tagExists := traffic.untagRevision(tag, serviceName)
		if !tagExists {
			errTagNames = append(errTagNames, tag)
		}
	}

	// Return all errors from untagging revisions
	if len(errTagNames) > 0 {
		return nil, fmt.Errorf("tag(s) %s not present for any revisions of service %s", strings.Join(errTagNames, ", "), serviceName)
	}

	for _, each := range trafficFlags.RevisionsTags {
		revision, tag, _ := splitByEqualSign(each) // err is checked in verifyInputSanity

		// Second precedence: Tag latestRevision
		if revision == latestRevisionRef {
			existingTagOnLatestRevision := traffic.tagOfLatestReadyRevision()

			// just pass if existing == requested
			if existingTagOnLatestRevision == tag {
				continue
			}

			// apply requested tag only if it doesnt exist in traffic block
			if traffic.isTagPresent(tag) {
				return nil, errorOverWritingTag(tag)
			}

			if existingTagOnLatestRevision == "" {
				traffic = traffic.TagLatestRevision(tag)
				continue
			} else {
				return nil, errorOverWritingtagOfLatestReadyRevision(existingTagOnLatestRevision, tag)
			}

		}

		// Third precedence: Tag other revisions
		// dont throw error if the tag present == requested tag
		if traffic.isTagPresentOnRevision(tag, revision) {
			continue
		}

		// error if the tag is assigned to some other revision
		if traffic.isTagPresent(tag) {
			return nil, errorOverWritingTag(tag)
		}

		traffic = traffic.TagRevision(tag, revision)
	}

	if cmd.Flags().Changed("traffic") {
		// reset existing traffic portions as what's on CLI is desired state of traffic split portions
		traffic.ResetAllTargetPercent()

		for _, each := range trafficFlags.RevisionsPercentages {
			// revisionRef works here as either revision or tag as either can be specified on CLI
			revisionRef, percent, _ := splitByEqualSign(each)  // err is verified in verifyInputSanity
			percentInt, _ := strconv.ParseInt(percent, 10, 64) // percentInt (for int) is verified in verifyInputSanity

			// fourth precedence: set traffic for latest revision
			if revisionRef == latestRevisionRef {
				if traffic.isLatestRevisionTrue() {
					traffic.SetTrafficByLatestRevision(percentInt)
				} else {
					// if no latestRevision ref is present in traffic block
					traffic = append(traffic, newTarget("", "", percentInt, true))
				}
				continue
			}

			// fifth precedence: set traffic for rest of revisions
			// If in a traffic block, revisionName of one target == tag of another,
			// one having tag is assigned given percent, as tags are supposed to be unique
			// and should be used (in this case) to avoid ambiguity

			// first check if given revisionRef is a tag
			if traffic.isTagPresent(revisionRef) {
				traffic.SetTrafficByTag(revisionRef, percentInt)
				continue
			}

			// check if given revisionRef is a revision
			if traffic.isRevisionPresent(revisionRef) {
				traffic.SetTrafficByRevision(revisionRef, percentInt)
				continue
			}

			// TODO Check at serving level, improve error
			//if !RevisionExists(revisionRef) {
			//	return error.New("Revision/Tag %s does not exists in traffic block.")
			//}

			// provided revisionRef isn't present in traffic block, add it
			traffic = append(traffic, newTarget("", revisionRef, percentInt, false))
		}
	}
	// remove any targets having no tags and 0% traffic portion
	return traffic.RemoveNullTargets(), nil
}
