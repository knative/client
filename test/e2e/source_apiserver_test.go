// Copyright 2019 The Knative Authors

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or im
// See the License for the specific language governing permissions and
// limitations under the License.

// +build e2e
// +build !serving

package e2e

import (
	"fmt"
	"testing"

	"gotest.tools/assert"
	"knative.dev/client/pkg/util"
)

func TestSourceApiServer(t *testing.T) {
	t.Parallel()
	test := NewE2eTest(t)
	test.Setup(t)
	defer test.Teardown(t)

	test.setupServiceAccountForApiserver(t, "mysa")
	test.serviceCreate(t, "myservice")

	t.Run("create apiserver source with a sink to a service", func(t *testing.T) {
		test.apiServerSourceCreate(t, "firstsrc", "Eventing:v1:true", "mysa", "svc:myservice")
		test.apiServerSourceCreate(t, "secondsrc", "Eventing,Namespace", "mysa", "svc:myservice")
	})

	t.Run("create apiserver source and delete it", func(t *testing.T) {
		test.apiServerSourceDelete(t, "firstsrc")
		test.apiServerSourceDelete(t, "secondsrc")
	})

	t.Run("create apiserver source with a missing sink service", func(t *testing.T) {
		test.apiServerSourceCreateMissingSink(t, "wrongsrc", "Eventing:v1:true", "mysa", "svc:unknown")
	})
}

func (test *e2eTest) apiServerSourceCreate(t *testing.T, srcName string, resources string, sa string, sink string) {
	out, err := test.kn.RunWithOpts([]string{"source", "apiserver", "create", srcName,
		"--resource", resources, "--service-account", sa, "--sink", sink}, runOpts{NoNamespace: false})
	assert.NilError(t, err)
	assert.Check(t, util.ContainsAllIgnoreCase(out, "ApiServerSource", srcName, "created", "namespace", test.kn.namespace))
}

func (test *e2eTest) apiServerSourceCreateMissingSink(t *testing.T, srcName string, resources string, sa string, sink string) {
	_, err := test.kn.RunWithOpts([]string{"source", "apiserver", "create", srcName,
		"--resource", resources, "--service-account", sa, "--sink", sink}, runOpts{NoNamespace: false, AllowError: true})
	assert.ErrorContains(t, err, "services.serving.knative.dev", "not found")
}

func (test *e2eTest) apiServerSourceDelete(t *testing.T, srcName string) {
	out, err := test.kn.RunWithOpts([]string{"source", "apiserver", "delete", srcName}, runOpts{NoNamespace: false})
	assert.NilError(t, err)
	assert.Check(t, util.ContainsAllIgnoreCase(out, "ApiServerSource", srcName, "deleted", "namespace", test.kn.namespace))
}

func (test *e2eTest) setupServiceAccountForApiserver(t *testing.T, name string) {
	kubectl := kubectl{t, Logger{}}

	_, err := kubectl.RunWithOpts([]string{"create", "serviceaccount", name}, runOpts{})
	if err != nil {
		t.Fatalf(fmt.Sprintf("Error executing 'kubectl create serviceaccount test-sa'. Error: %s", err.Error()))
	}
	_, err = kubectl.RunWithOpts([]string{"create", "clusterrole", "testsa-role", "--verb=get,list,watch", "--resource=events,namespaces"}, runOpts{})
	if err != nil {
		t.Fatalf(fmt.Sprintf("Error executing 'kubectl clusterrole testsa-role'. Error: %s", err.Error()))
	}
	_, err = kubectl.RunWithOpts([]string{"create", "clusterrolebinding", "testsa-binding", "--clusterrole=testsa-role", "--serviceaccount=" + test.kn.namespace + ":" + name}, runOpts{})
	if err != nil {
		t.Fatalf(fmt.Sprintf("Error executing 'kubectl clusterrolebinding testsa-binding'. Error: %s", err.Error()))
	}
}
