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
)

// Default time out to use when waiting for reconciliation. It is deliberately very long as it is expected that
// the service doesn't stay in `Unknown` status very long and eventually ends up as `False` or `True` in a timely
// manner
const WaitDefaultTimeout = 600

// Flags for tuning wait behaviour
type WaitFlags struct {
	// Timeout in seconds for how long to wait for a command to return
	TimeoutInSeconds int

	// If set then just apply resources and don't wait
	NoWait bool
	//TODO: deprecated variable should be removed with --async flag
	Async bool
}

// Add flags which influence the sync/async behaviour when creating or updating
// resources. Set `waitDefault` argument if the default behaviour is synchronous.
// Use `what` for describing what is waited for.
func (p *WaitFlags) AddConditionWaitFlags(command *cobra.Command, waitTimeoutDefault int, action, what, until string) {
	waitUsage := fmt.Sprintf("%s %s and don't wait for it to be %s.", action, what, until)
	//TODO: deprecated flag should be removed in next release
	command.Flags().BoolVar(&p.Async, "async", false, "DEPRECATED: please use --no-wait instead. "+waitUsage)
	command.Flags().BoolVar(&p.NoWait, "no-wait", false, waitUsage)

	timeoutUsage := fmt.Sprintf("Seconds to wait before giving up on waiting for %s to be %s.", what, until)
	command.Flags().IntVar(&p.TimeoutInSeconds, "wait-timeout", waitTimeoutDefault, timeoutUsage)
}
