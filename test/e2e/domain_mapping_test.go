// Copyright 2021 The Knative Authors

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build e2e
// +build !eventing

package e2e

import (
	"testing"

	"knative.dev/client/pkg/util"

	"gotest.tools/v3/assert"

	"knative.dev/client/lib/test"
)

func TestDomain(t *testing.T) {
	t.Parallel()
	it, err := test.NewKnTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, it.Teardown())
	}()

	r := test.NewKnRunResultCollector(t, it)
	defer r.DumpIfFailed()

	domainName := "hello.example.com"

	t.Log("create domain mapping to hello ksvc")
	test.ServiceCreate(r, "hello")
	domainCreate(r, domainName, "hello")

	t.Log("list domain mappings")
	domainList(r, domainName)

	t.Log("update domain mapping Knative service reference")
	test.ServiceCreate(r, "foo")
	domainUpdate(r, domainName, "foo")

	t.Log("describe domain mappings")
	domainDescribe(r, domainName, false)

	t.Log("delete domain")
	domainDelete(r, domainName)

	t.Log("create domain with TLS")
	domainCreateWithTls(r, "newdomain.com", "hello", "tls-secret")
	domainDescribe(r, "newdomain.com", true)
}

func domainCreate(r *test.KnRunResultCollector, domainName, serviceName string, options ...string) {
	command := []string{"domain", "create", domainName, "--ref", serviceName}
	command = append(command, options...)
	out := r.KnTest().Kn().Run(command...)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "domain", "mapping", serviceName, domainName, "created", "namespace", r.KnTest().Kn().Namespace()))
}

func domainCreateWithTls(r *test.KnRunResultCollector, domainName, serviceName, tls string, options ...string) {
	command := []string{"domain", "create", domainName, "--ref", serviceName, "--tls", tls}
	command = append(command, options...)
	out := r.KnTest().Kn().Run(command...)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "domain", "mapping", serviceName, domainName, "created", "namespace", r.KnTest().Kn().Namespace()))
}

func domainUpdate(r *test.KnRunResultCollector, domainName, serviceName string, options ...string) {
	command := []string{"domain", "update", domainName, "--ref", serviceName}
	command = append(command, options...)
	out := r.KnTest().Kn().Run(command...)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "domain", "mapping", domainName, "updated", "namespace", r.KnTest().Kn().Namespace()))
}

func domainDelete(r *test.KnRunResultCollector, domainName string, options ...string) {
	command := []string{"domain", "delete", domainName}
	command = append(command, options...)
	out := r.KnTest().Kn().Run(command...)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "domain", "mapping", domainName, "deleted", "namespace", r.KnTest().Kn().Namespace()))
}

func domainList(r *test.KnRunResultCollector, domainName string) {
	out := r.KnTest().Kn().Run("domain", "list")
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stdout, domainName))
}

func domainDescribe(r *test.KnRunResultCollector, domainName string, tls bool) {
	out := r.KnTest().Kn().Run("domain", "describe", domainName)
	r.AssertNoError(out)
	var url string
	if tls {
		url = "https://" + domainName
	} else {
		url = "http://" + domainName
	}
	assert.Assert(r.T(), util.ContainsAll(out.Stdout, "Name", "Namespace", "URL", "Service", url))
}
