/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package test

import (
	"time"

	"gotest.tools/assert"

	"knative.dev/client/pkg/util"
)

func SubscriptionCreate(r *KnRunResultCollector, sname string, args ...string) {
	cmd := []string{"subscription", "create", sname}
	cmd = append(cmd, args...)
	out := r.KnTest().Kn().Run(cmd...)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "subscription", sname, "created"))
	// let the subscription and related resource reconcile
	time.Sleep(time.Second * 5)
}

func SubscriptionList(r *KnRunResultCollector, args ...string) string {
	cmd := []string{"subscription", "list"}
	cmd = append(cmd, args...)
	out := r.KnTest().Kn().Run(cmd...)
	r.AssertNoError(out)
	return out.Stdout
}

func SubscriptionDescribe(r *KnRunResultCollector, sname string, args ...string) string {
	cmd := []string{"subscription", "describe", sname}
	cmd = append(cmd, args...)
	out := r.KnTest().Kn().Run(cmd...)
	r.AssertNoError(out)
	return out.Stdout
}

func SubscriptionDelete(r *KnRunResultCollector, sname string) {
	out := r.KnTest().Kn().Run("subscription", "delete", sname)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "subscription", sname, "deleted"))
}

func SubscriptionUpdate(r *KnRunResultCollector, sname string, args ...string) {
	cmd := []string{"subscription", "update", sname}
	cmd = append(cmd, args...)
	out := r.KnTest().Kn().Run(cmd...)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "subscription", sname, "updated"))
	// let the subscription and related resource reconcile
	time.Sleep(time.Second * 5)
}
