// Copyright © 2022 The Knative Authors
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

package broker

import (
	"github.com/spf13/cobra"
	"knative.dev/client/pkg/kn/commands/flags"
)

type DeliveryOptionFlags struct {
	SinkFlags     flags.SinkFlags
	RetryCount    int32
	Timeout       string
	BackoffPolicy string
	BackoffDelay  string
	RetryAfterMax string
}

func (d *DeliveryOptionFlags) Add(cmd *cobra.Command) {
	d.SinkFlags.AddWithFlagName(cmd, "dl-sink", "")
	cmd.Flag("dl-sink").Usage = "Reference to a sink for delivering events that can not be sent"

	cmd.Flags().Int32Var(&d.RetryCount, "retry", 0, "Number of retries before sending the event to a dead-letter sink")
	cmd.Flags().StringVar(&d.Timeout, "timeout", "", "Timeout for a single request")
	cmd.Flags().StringVar(&d.BackoffPolicy, "backoff-policy", "", "Backoff policy for retries, either \"linear\" or \"exponential\"")
	cmd.Flags().StringVar(&d.BackoffDelay, "backoff-delay", "", "Based delay between retries")
	cmd.Flags().StringVar(&d.RetryAfterMax, "retry-after-max", "", "Upper bound for a duration specified in an \"Retry-After\" header (experimental)")
}
