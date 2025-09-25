// Copyright © 2021 The Knative Authors
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

package quickstart

import (
	"fmt"

	"github.com/spf13/cobra"
)

var name string
var kubernetesVersion string
var installServing bool
var installEventing bool
var installKindRegistry bool
var installKindExtraMountHostPath string
var installKindExtraMountContainerPath string

func clusterNameOption(targetCmd *cobra.Command, flagDefault string) {
	targetCmd.Flags().StringVarP(
		&name,
		"name",
		"n",
		flagDefault,
		fmt.Sprintf("%s cluster name to be used by kn quickstart", targetCmd.Name()),
	)
}

func kubernetesVersionOption(targetCmd *cobra.Command, flagDefault string, usageText string) {
	targetCmd.Flags().StringVarP(
		&kubernetesVersion,
		"kubernetes-version",
		"k",
		flagDefault,
		usageText)
}

func installServingOption(targetCmd *cobra.Command) {
	targetCmd.Flags().BoolVar(&installServing, "install-serving", false, "install Serving on quickstart cluster")
}

func installEventingOption(targetCmd *cobra.Command) {
	targetCmd.Flags().BoolVar(&installEventing, "install-eventing", false, "install Eventing on quickstart cluster")
}

func installKindRegistryOption(targetCmd *cobra.Command) {
	targetCmd.Flags().BoolVar(&installKindRegistry, "registry", false, "install registry for Kind quickstart cluster")
}

func installKindExtraMountHostPathOption(targetCmd *cobra.Command) {
	targetCmd.Flags().StringVarP(&installKindExtraMountHostPath, "extraMountHostPath", "", "", "set the extraMount hostPath on Kind quickstart cluster")
}

func installKindExtraMountContainerPathOption(targetCmd *cobra.Command) {
	targetCmd.Flags().StringVarP(&installKindExtraMountContainerPath, "extraMountContainerPath", "", "", "set the extraMount containerPath on Kind quickstart cluster")
}