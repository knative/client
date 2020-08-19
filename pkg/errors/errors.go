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

package errors

import (
	"fmt"
	"strings"
)

func newInvalidCRD(apiGroup string) *KNError {
	parts := strings.Split(apiGroup, ".")
	name := parts[0]
	msg := fmt.Sprintf("no Knative %s API found on the backend, please verify the installation", name)
	return NewKNError(msg)
}

func newNoRouteToHost(errString string) *KNError {
	parts := strings.SplitAfter(errString, "dial tcp")
	if len(parts) == 2 {
		return NewKNError(fmt.Sprintf("error connecting to the cluster, please verify connection at: %s", strings.Trim(parts[1], " ")))
	}
	return NewKNError(fmt.Sprintf("error connecting to the cluster: %s", errString))
}

func newNoKubeConfig(errString string) *KNError {
	return NewKNError("no kubeconfig has been provided, please use a valid configuration to connect to the cluster")
}
