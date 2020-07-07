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
	"knative.dev/eventing/pkg/apis/eventing/v1beta1"
)

// LabelNamespaceForDefaultBroker adds label 'knative-eventing-injection=enabled' to the configured namespace
func LabelNamespaceForDefaultBroker(r *KnRunResultCollector) error {
	cmd := []string{"label", "namespace", r.KnTest().Kn().Namespace(), v1beta1.InjectionAnnotation + "=enabled"}
	_, err := Kubectl{}.Run(cmd...)

	if err != nil {
		r.T().Fatalf("error executing '%s': %s", strings.Join(cmd, " "), err.Error())
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
	cmd := []string{"label", "namespace", r.KnTest().Kn().Namespace(), v1beta1.InjectionAnnotation + "-"}
	_, err := Kubectl{}.Run(cmd...)
	if err != nil {
		r.T().Fatalf("error executing '%s': %s", strings.Join(cmd, " "), err.Error())
	}
}
