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

package flags

import "github.com/spf13/cobra"

type EventtypeFlags struct {
	Type   string
	Source string
	Broker string
}

func (e *EventtypeFlags) Add(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&e.Type, "type", "t", "", "Cloud Event type")
	cmd.Flags().StringVar(&e.Source, "source", "", "Cloud Event source")
	cmd.Flags().StringVarP(&e.Broker, "broker", "b", "", "Cloud Event Broker. This flag is added for the convenience, since Eventing v1beta2 brokers as represented as KReference type.")
	cmd.MarkFlagRequired("type")
}
