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
	"reflect"

	"github.com/spf13/cobra"
	sinkfl "knative.dev/client-pkg/pkg/commands/flags/sink"
	"knative.dev/client-pkg/pkg/dynamic"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

type DeliveryOptionFlags struct {
	SinkFlags     sinkfl.Flag
	RetryCount    int32
	Timeout       string
	BackoffPolicy string
	BackoffDelay  string
	RetryAfterMax string
}

func (d *DeliveryOptionFlags) Add(cmd *cobra.Command) {
	d.SinkFlags.AddWithFlagName(cmd, "dl-sink", "")
	cmd.Flag("dl-sink").Usage = "The sink receiving event that could not be sent to a destination."

	cmd.Flags().Int32Var(&d.RetryCount, "retry", 0, "The minimum number of retries the sender should attempt when "+
		"sending an event before moving it to the dead letter sink.")
	cmd.Flags().StringVar(&d.Timeout, "timeout", "", "The timeout of each single request. The value must be greater than 0.")
	cmd.Flags().StringVar(&d.BackoffPolicy, "backoff-policy", "", "The retry backoff policy (linear, exponential).")
	cmd.Flags().StringVar(&d.BackoffDelay, "backoff-delay", "", "The delay before retrying.")
	cmd.Flags().StringVar(&d.RetryAfterMax, "retry-after-max", "", "An optional upper bound on the duration specified in a "+
		"\"Retry-After\" header when calculating backoff times for retrying 429 and 503 response codes. "+
		"Setting the value to zero (\"PT0S\") can be used to opt-out of respecting \"Retry-After\" header values altogether. "+
		"This value only takes effect if \"Retry\" is configured, and also depends on specific implementations (Channels, Sources, etc.) "+
		"choosing to provide this capability.")
}

func (d *DeliveryOptionFlags) GetDlSink(cmd *cobra.Command, dynamicClient dynamic.KnDynamicClient, namespace string) (*duckv1.Destination, error) {
	var empty = sinkfl.Flag{}
	var destination *duckv1.Destination
	var err error
	if !reflect.DeepEqual(d.SinkFlags, empty) {
		destination, err = d.SinkFlags.ResolveSink(cmd.Context(), dynamicClient, namespace)
		if err != nil {
			return nil, err
		}
	}
	return destination, err
}
