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

package flags

import (
	"github.com/spf13/cobra"
)

type Traffic struct {
	RevisionsPercentages []string
	RevisionsTags        []string
	UntagRevisions       []string
}

func (t *Traffic) Add(cmd *cobra.Command) {
	cmd.Flags().StringSliceVar(&t.RevisionsPercentages,
		"traffic",
		nil,
		"Set traffic distribution (format: --traffic revisionRef=percent) where revisionRef can be a revision or a tag or '@latest' string "+
			"representing latest ready revision. This flag can be given multiple times with percent summing up to 100%.")

	cmd.Flags().StringSliceVar(&t.RevisionsTags,
		"tag",
		nil,
		"Set tag (format: --tag revisionRef=tagName) where revisionRef can be a revision or '@latest' string representing latest ready revision. "+
			"This flag can be specified multiple times.")

	cmd.Flags().StringSliceVar(&t.UntagRevisions,
		"untag",
		nil,
		"Untag revision (format: --untag tagName). This flag can be specified multiple times.")
}

func (t *Traffic) PercentagesChanged(cmd *cobra.Command) bool {
	return cmd.Flags().Changed("traffic")
}

func (t *Traffic) TagsChanged(cmd *cobra.Command) bool {
	return cmd.Flags().Changed("tag") || cmd.Flags().Changed("untag")
}

func (t *Traffic) Changed(cmd *cobra.Command) bool {
	return t.PercentagesChanged(cmd) || t.TagsChanged(cmd)
}
