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

package channel

import (
	"github.com/spf13/cobra"
	"knative.dev/client/pkg/kn/config"
)

type ChannelTypeMappingFlags struct {
	config.ChannelTypeMapping
}

func (ctmf *ChannelTypeMappingFlags) AddFlags(c *cobra.Command) {
	c.Flags().StringVar(&ctmf.Kind, "kind", "", "The kind of Channel, such as InMemoryChannel or KafkaChannel, based on the default ConfigMap")
	c.Flags().StringVar(&ctmf.Group, "group", "", "The API group of the Kubernetes resource")
	c.Flags().StringVar(&ctmf.Version, "version", "", "The version of the Kubernetes resource")
}
