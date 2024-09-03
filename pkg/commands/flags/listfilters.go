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
	"fmt"

	"github.com/spf13/cobra"
)

// SourceTypeFilters defines flags used for kn source list to filter sources on types
type SourceTypeFilters struct {
	Filters []string
}

// Add attaches the SourceTypeFilters flags to given command
func (s *SourceTypeFilters) Add(cmd *cobra.Command, what string) {
	usage := fmt.Sprintf("Filter list on given %s. This flag can be given multiple times.", what)
	cmd.Flags().StringSliceVarP(&s.Filters, "type", "t", nil, usage)
}
