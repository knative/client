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
	"github.com/spf13/cobra"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"knative.dev/client/pkg/kn/commands"
)

func NewRevisionCommand(p *commands.KnParams) *cobra.Command {
	revisionCmd := &cobra.Command{
		Use:   "revision",
		Short: "Revision command group",
	}
	revisionCmd.AddCommand(NewRevisionListCommand(p))
	revisionCmd.AddCommand(NewRevisionDescribeCommand(p))
	revisionCmd.AddCommand(NewRevisionDeleteCommand(p))
	return revisionCmd
}

// ============================================
// Shared revision functions:

// Extract traffic and tags for given revision from a service
func trafficAndTagsForRevision(revision string, service *servingv1.Service) (int64, []string) {
	if len(service.Status.Traffic) == 0 {
		return 0, nil
	}
	var percent int64
	var tags []string
	for _, target := range service.Status.Traffic {
		if target.RevisionName == revision {
			if target.Percent != nil {
				percent += *target.Percent
			}
			if target.Tag != "" {
				tags = append(tags, target.Tag)
			}
		}
	}
	return percent, tags
}
