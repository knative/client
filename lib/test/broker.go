// Copyright 2020 The Knative Authors

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or im
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

// LabelNamespaceForDefaultBroker adds label 'knative-eventing-injection=enabled' to the configured namespace
func LabelNamespaceForDefaultBroker(r *KnRunResultCollector) error {
	_, err := Kubectl{}.Run("label", "namespace", r.KnTest().Kn().Namespace(), "knative-eventing-injection=enabled")

	if err != nil {
		r.T().Fatalf("Error executing 'kubectl label namespace %s knative-eventing-injection=enabled'. Error: %s", r.KnTest().Kn().Namespace(), err.Error())
	}

	return wait.PollImmediate(10*time.Second, 5*time.Minute, func() (bool, error) {
		out, err := NewKubectl(r.KnTest().Kn().Namespace()).Run("get", "broker", "-o=jsonpath='{.items[0].status.conditions[?(@.type==\"Ready\")].status}'")
		if err != nil {
			return false, nil
		}

		return strings.Contains(out, "True"), nil
	})
}

// UnlabelNamespaceForDefaultBroker removes label 'knative-eventing-injection=enabled' from the configured namespace
func UnlabelNamespaceForDefaultBroker(r *KnRunResultCollector) {
	_, err := Kubectl{}.Run("label", "namespace", r.KnTest().Kn().Namespace(), "knative-eventing-injection-")
	if err != nil {
		r.T().Fatalf("Error executing 'kubectl label namespace %s knative-eventing-injection-'. Error: %s", r.KnTest().Kn().Namespace(), err.Error())
	}
}
