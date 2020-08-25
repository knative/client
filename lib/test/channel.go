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
	"time"

	"gotest.tools/assert"

	"knative.dev/client/pkg/util"
)

func ChannelCreate(r *KnRunResultCollector, cname string, args ...string) {
	cmd := []string{"channel", "create", cname}
	cmd = append(cmd, args...)
	out := r.KnTest().Kn().Run(cmd...)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "channel", cname, "created"))
	// let channel reconcile TODO: fix the wait for channel to become ready
	time.Sleep(5 * time.Second)
}

func ChannelList(r *KnRunResultCollector, args ...string) string {
	cmd := []string{"channel", "list"}
	cmd = append(cmd, args...)
	out := r.KnTest().Kn().Run(cmd...)
	r.AssertNoError(out)
	return out.Stdout
}

func ChannelDescribe(r *KnRunResultCollector, cname string, args ...string) string {
	cmd := []string{"channel", "describe", cname}
	cmd = append(cmd, args...)
	out := r.KnTest().Kn().Run(cmd...)
	r.AssertNoError(out)
	return out.Stdout
}

func ChannelDelete(r *KnRunResultCollector, cname string) {
	out := r.KnTest().Kn().Run("channel", "delete", cname)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "channel", cname, "deleted"))
}
