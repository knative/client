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

package test

import (
	"gotest.tools/v3/assert"
	"knative.dev/client/pkg/util"
)

// EventtypeCreate creates an eventtype with the given name.
func EventtypeCreate(r *KnRunResultCollector, name, cetype string) {
	out := r.KnTest().Kn().Run("eventtype", "create", name, "--type", cetype)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "Eventtype", name, "created", "namespace", r.KnTest().Kn().Namespace()))
}

// EventtypeDelete deletes an eventtype with the given name.
func EventtypeDelete(r *KnRunResultCollector, name string) {
	out := r.KnTest().Kn().Run("eventtype", "delete", name)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "Eventtype", name, "deleted", "namespace", r.KnTest().Kn().Namespace()))
}

// EventtypeList verifies listing eventtypes in the given namespace
func EventtypeList(r *KnRunResultCollector, eventtypes ...string) {
	out := r.KnTest().Kn().Run("eventtype", "list")
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, eventtypes...))
}

// EventtypeDescribe describes an eventtype with the given name.
func EventtypeDescribe(r *KnRunResultCollector, name string) {
	out := r.KnTest().Kn().Run("eventtype", "describe", name)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stdout, name, r.KnTest().Kn().Namespace(), "Conditions"))
}

func EventtypeCreateWithBrokerSource(r *KnRunResultCollector, name, cetype, broker, source string) {
	out := r.KnTest().Kn().Run("eventtype", "create", name, "--type", cetype, "--broker", broker, "--source", source)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "Eventtype", name, "created", "namespace", r.KnTest().Kn().Namespace()))
}

func EventtypeCreateWithSourceError(r *KnRunResultCollector, name, cetype, source string) {
	out := r.KnTest().Kn().Run("eventtype", "create", name, "--type", cetype, "--source", source)
	r.AssertError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stderr, name, "invalid", "control character"))
}
