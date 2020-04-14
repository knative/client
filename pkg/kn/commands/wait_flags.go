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

package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	knflags "knative.dev/client/pkg/kn/flags"
)

// Default time out to use when waiting for reconciliation. It is deliberately very long as it is expected that
// the service doesn't stay in `Unknown` status very long and eventually ends up as `False` or `True` in a timely
// manner
const WaitDefaultTimeout = 600

// Flags for tuning wait behaviour
type WaitFlags struct {
	// Timeout in seconds for how long to wait for a command to return
	TimeoutInSeconds int
	// If set then apply resources and wait for completion
	Wait bool
	//TODO: deprecated variable should be removed with --async flag
	Async bool
}

// Add flags which influence the sync/async behaviour when creating or updating
// resources. Set `waitDefault` argument if the default behaviour is synchronous.
// Use `what` for describing what is waited for.
func (p *WaitFlags) AddConditionWaitFlags(command *cobra.Command, waitTimeoutDefault int, action, what, until string) {
	waitUsage := fmt.Sprintf("Wait for '%s %s' operation to be completed.", what, action)
	waitDefault := true
	// Special-case 'delete' command so it comes back to the user immediately
	if action == "delete" {
		waitDefault = false
	}

	//TODO: deprecated flag should be removed in next release
	command.Flags().BoolVar(&p.Async, "async", !waitDefault, "DEPRECATED: please use --no-wait instead. "+knflags.InvertUsage(waitUsage))
	knflags.AddBothBoolFlagsUnhidden(command.Flags(), &p.Wait, "wait", "", waitDefault, waitUsage)
	timeoutUsage := fmt.Sprintf("Seconds to wait before giving up on waiting for %s to be %s.", what, until)
	command.Flags().IntVar(&p.TimeoutInSeconds, "wait-timeout", waitTimeoutDefault, timeoutUsage)
}
