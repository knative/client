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

	"gotest.tools/v3/assert"
	"knative.dev/client/pkg/util"

	"k8s.io/apimachinery/pkg/util/wait"
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
)

// BrokerCreate creates a broker with the given name.
func BrokerCreate(r *KnRunResultCollector, name string) {
	out := r.KnTest().Kn().Run("broker", "create", name)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "Broker", name, "created", "namespace", r.KnTest().Kn().Namespace()))
}

// BrokerCreateWithClass creates a broker with the given name and class.
func BrokerCreateWithClass(r *KnRunResultCollector, name, class string) {
	out := r.KnTest().Kn().Run("broker", "create", name, "--class", class)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "Broker", name, "created", "namespace", r.KnTest().Kn().Namespace()))
}

// BrokerDelete deletes a broker with the given name.
func BrokerDelete(r *KnRunResultCollector, name string, wait bool) {
	args := []string{"broker", "delete", name}
	if wait {
		args = append(args, "--wait")
	}
	out := r.KnTest().Kn().Run(args...)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "Broker", name, "deleted", "namespace", r.KnTest().Kn().Namespace()))
}

// LabelNamespaceForDefaultBroker adds label 'knative-eventing-injection=enabled' to the configured namespace
func LabelNamespaceForDefaultBroker(r *KnRunResultCollector) error {
	cmd := []string{"label", "namespace", r.KnTest().Kn().Namespace(), eventingv1.InjectionAnnotation + "=enabled"}
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
	cmd := []string{"label", "namespace", r.KnTest().Kn().Namespace(), eventingv1.InjectionAnnotation + "-"}
	_, err := Kubectl{}.Run(cmd...)
	if err != nil {
		r.T().Fatalf("error executing '%s': %s", strings.Join(cmd, " "), err.Error())
	}
}
