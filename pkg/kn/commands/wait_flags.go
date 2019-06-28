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

// Flags for tuning wait behaviour
type WaitFlags struct {
	// Timeout in seconds for how long to wait for a command to return
	TimeoutInSeconds int

	// If set then just apply resources and don't wait
	Async bool
}

// Add flags which influence the sync/async behaviour when creating or updating
// resources. Set `waitDefault` argument if the default behaviour is synchronous.
// Use `what` for describing what is waited for.
func (p *WaitFlags) AddWaitFlags(command *cobra.Command, waitTimeoutDefault int, what string) {
	waitUsage := fmt.Sprintf("Create %s and don't wait for it to become ready.", what)
	command.Flags().BoolVar(&p.Async, "async", false, waitUsage)

	timeoutUsage := fmt.Sprintf("Seconds to wait before giving up on waiting for %s to be ready (default: %d).", what, waitTimeoutDefault)
	command.Flags().IntVar(&p.TimeoutInSeconds, "wait-timeout", waitTimeoutDefault, timeoutUsage)
}
