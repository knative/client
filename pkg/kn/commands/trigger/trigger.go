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

package trigger

import (
	"github.com/spf13/cobra"

	"knative.dev/client/pkg/kn/commands"
)

const (
	// How often to retry in case of an optimistic lock error when replacing a trigger (--force)
	MaxUpdateRetries = 3
)

// NewTriggerCommand to create trigger command group
func NewTriggerCommand(p *commands.KnParams) *cobra.Command {
	triggerCmd := &cobra.Command{
		Use:   "trigger",
		Short: "Trigger command group",
	}
	triggerCmd.AddCommand(NewTriggerCreateCommand(p))
	triggerCmd.AddCommand(NewTriggerUpdateCommand(p))
	triggerCmd.AddCommand(NewTriggerDescribeCommand(p))
	triggerCmd.AddCommand(NewTriggerListCommand(p))
	triggerCmd.AddCommand(NewTriggerDeleteCommand(p))
	return triggerCmd
}
