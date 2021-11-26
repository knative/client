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
	"knative.dev/client/pkg/kn/commands/revision"
	"knative.dev/pkg/ptr"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"knative.dev/client/pkg/kn/commands/flags"
)

var latestRevisionRef = "@latest"

const (
	errorDistributionRevisionCount = iota
	errorDistributionRevisionNotFound
)

// ServiceTraffic type for operating on service traffic targets
type ServiceTraffic []servingv1.TrafficTarget

func newServiceTraffic(traffic []servingv1.TrafficTarget) ServiceTraffic {
	return ServiceTraffic(traffic)
}

func splitByEqualSign(pair string) (string, string, error) {
	parts := strings.Split(pair, "=")
	if len(parts) == 1 {
		return latestRevisionRef, parts[0], nil
	}
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

func (e ServiceTraffic) untagRevision(tag string) bool {
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

func errorTrafficDistribution(sum int, reason int) error {
	errMsg := ""
	switch reason {
	case errorDistributionRevisionCount:
		errMsg = "incorrect number of revisions specified. Only 1 revision should be missing"
	case errorDistributionRevisionNotFound:
		errMsg = "cannot determine the missing revision"
	}
	return fmt.Errorf("unable to allocate the remaining traffic: %d%%. Reason: %s", 100-sum, errMsg)
}

func errorSumGreaterThan100(sum int) error {
	return fmt.Errorf("given traffic percents sum to %d, want <=100", sum)
}

// verifies if user has repeated @latest field in --tag or --traffic flags
// verifyInput checks:
// - if user has repeated @latest field in --tag or --traffic flags
// - if provided traffic portion are integers
// - if traffic as per flags sums to 100
// - if traffic as per flags < 100, can remaining traffic be automatically directed
func verifyInput(trafficFlags *flags.Traffic, svc *servingv1.Service, revisions []servingv1.Revision, mutation bool) error {
	// check if traffic is being sent to @latest tag
	var latestRefTraffic bool

	// number of existing revisions
	var existingRevisionCount = len(revisions)

	err := verifyLatestTag(trafficFlags)
	if err != nil {
		return err
	}

	revisionRefMap, sum, err := verifyRevisionSumAndReferences(trafficFlags)
	if err != nil {
		return err
	}

	// further verification is not required if sum >= 100
	if sum == 100 {
		return nil
	}
	if sum > 100 {
		return errorSumGreaterThan100(sum)
	}

	if _, ok := revisionRefMap[latestRevisionRef]; ok {
		// traffic has been routed to @latest tag
		latestRefTraffic = true
	}

	// number of revisions specified in traffic flags
	specRevPercentCount := len(trafficFlags.RevisionsPercentages)

	// no traffic to route
	if specRevPercentCount == 0 {
		return nil
	}

	// cannot determine remaining revision. Only 1 should be left out
	if specRevPercentCount < existingRevisionCount-1 {
		return errorTrafficDistribution(sum, errorDistributionRevisionCount)
	}

	if specRevPercentCount == existingRevisionCount-1 {
		if latestRefTraffic {
			// if mutation is set, @latest tag is referring to the revision which
			// will be created after service update. specRevPercentCount should have been
			// equal to existingRevisionCount
			if mutation {
				return errorTrafficDistribution(sum, errorDistributionRevisionCount)
			}
			revisionRefMap[svc.Status.LatestReadyRevisionName] = 0
		}
	}

	if specRevPercentCount == existingRevisionCount && latestRefTraffic {
		// if mutation is not set, @latest tag is referring to the revision which
		// already exists. specRevPercentCount should have been equal to existingRevisionCount
		if !mutation {
			return errorTrafficDistribution(sum, errorDistributionRevisionCount)
		}
	}

	// remaining % to 100
	for _, rev := range revisions {
		if !checkRevisionPresent(revisionRefMap, rev) {
			trafficFlags.RevisionsPercentages = append(trafficFlags.RevisionsPercentages, fmt.Sprintf("%s=%d", rev.Name, 100-sum))
			return nil
		}
	}

	return errorTrafficDistribution(sum, errorDistributionRevisionNotFound)
}

func verifyRevisionSumAndReferences(trafficFlags *flags.Traffic) (revisionRefMap map[string]int, sum int, err error) {

	revisionRefMap = make(map[string]int)
	for i, each := range trafficFlags.RevisionsPercentages {
		var revisionRef, percent string
		revisionRef, percent, err = splitByEqualSign(each)
		if err != nil {
			return
		}
		// To check if there are duplicate revision names in traffic flags
		if _, exist := revisionRefMap[revisionRef]; exist {
			err = errorRepeatingRevision("--traffic", revisionRef)
			return
		}

		revisionRefMap[revisionRef] = i

		var percentInt int
		percentInt, err = strconv.Atoi(percent)
		if err != nil {
			err = errorParsingInteger(percent)
			return
		}

		if percentInt < 0 || percentInt > 100 {
			err = fmt.Errorf("invalid value for traffic percent %d, expected 0 <= percent <= 100", percentInt)
			return
		}

		sum += percentInt
	}
	return
}

func errorParsingInteger(percent string) error {
	return fmt.Errorf("error converting given %s to integer value for traffic distribution", percent)
}

func verifyLatestTag(trafficFlags *flags.Traffic) error {
	var latestRevisionTag bool
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
	return nil
}

func checkRevisionPresent(refMap map[string]int, rev servingv1.Revision) bool {
	_, nameExists := refMap[rev.Name]
	_, tagExists := refMap[rev.Annotations[revision.RevisionTagsAnnotation]]
	return tagExists || nameExists
}

// Compute takes service object, computes traffic per given traffic flags and returns the new traffic. If
// total traffic per flags < 100, the params 'revisions' and 'mutation' are used to direct remaining
// traffic. Param 'mutation' is set to true if a new revision will be created on service update
func Compute(cmd *cobra.Command, svc *servingv1.Service,
	trafficFlags *flags.Traffic, revisions []servingv1.Revision, mutation bool) ([]servingv1.TrafficTarget, error) {
	targets := svc.Spec.Traffic
	serviceName := svc.Name
	err := verifyInput(trafficFlags, svc, revisions, mutation)
	if err != nil {
		return nil, err
	}

	traffic := newServiceTraffic(targets)

	// First precedence: Untag revisions
	var errTagNames []string
	for _, tag := range trafficFlags.UntagRevisions {
		tagExists := traffic.untagRevision(tag)
		if !tagExists {
			errTagNames = append(errTagNames, tag)
		}
	}

	// Return all errors from untagging revisions
	if len(errTagNames) > 0 {
		return nil, fmt.Errorf("tag(s) %s not present for any revisions of service %s", strings.Join(errTagNames, ", "), serviceName)
	}

	for _, each := range trafficFlags.RevisionsTags {
		revision, tag, _ := splitByEqualSign(each) // err is checked in verifyInput

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

	if cmd.Flags().Changed("traffic") || (cmd.Name() == "create" && len(trafficFlags.RevisionsTags) > 0) {
		// reset existing traffic portions as what's on CLI is desired state of traffic split portions
		traffic.ResetAllTargetPercent()

		for _, each := range trafficFlags.RevisionsPercentages {
			// revisionRef works here as either revision or tag as either can be specified on CLI
			revisionRef, percent, _ := splitByEqualSign(each)  // err is verified in verifyInput
			percentInt, _ := strconv.ParseInt(percent, 10, 64) // percentInt (for int) is verified in verifyInput

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
