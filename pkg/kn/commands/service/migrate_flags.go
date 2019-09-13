// Copyright © 2018 The Knative Authors
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

package service

import (
	"github.com/spf13/cobra"
)

type MigrateFlags struct {
	// Migration information
	SourceKubeconfig, DestinationKubeconfig string
	SourceNamespace, DestinationNamespace   string

	// Preferences about how to do the action.
	ForceReplace   bool
	DeleteOriginal bool
}

// addFlags adds the flags to kn-migration.
func (p *MigrateFlags) addFlags(command *cobra.Command) {
	command.Flags().StringVar(&p.SourceKubeconfig, "source-kubeconfig", "", "The kubeconfig of the source Knative resources (default is KUBECONFIG2 from ENV property)")
	command.Flags().StringVar(&p.SourceNamespace, "source-namespace", "", "The namespace of the source Knative resources")

	command.Flags().StringVar(&p.DestinationKubeconfig, "destination-kubeconfig", "", "The kubeconfig of the destination Knative resources (default is KUBECONFIG2 from ENV property)")
	command.Flags().StringVar(&p.DestinationNamespace, "destination-namespace", "", "The namespace of the destination Knative resources")

	command.Flags().BoolVar(&p.ForceReplace, "force", false, "Migrate service forcefully, replaces existing service if any.")
	command.Flags().BoolVar(&p.DeleteOriginal, "delete", false, "Delete all Knative resources after kn-migration from source cluster")
}
