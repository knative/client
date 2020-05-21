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

// +build e2e
// +build !serving

package e2e

import (
	"io/ioutil"
	"os"
	"testing"

	"gotest.tools/assert"

	"knative.dev/client/lib/test"
	"knative.dev/client/pkg/util"
)

const (
	KnConfigContent string = `sink:
- group: serving.knative.dev
  prefix: hello
  resource: services
  version: v1`
)

type sinkprefixTestConfig struct {
	knConfigDir  string
	knConfigPath string
}

func (tc *sinkprefixTestConfig) setup() error {
	var err error
	tc.knConfigDir, err = ioutil.TempDir("", "kn1-config")
	if err != nil {
		return err
	}
	tc.knConfigPath, err = test.CreateFile("config.yaml", KnConfigContent, tc.knConfigDir, test.FileModeReadWrite)
	if err != nil {
		return err
	}
	return nil
}

func (tc *sinkprefixTestConfig) teardown() {
	os.RemoveAll(tc.knConfigDir)
}

func TestSinkPrefixConfig(t *testing.T) {
	t.Parallel()
	it, err := test.NewKnTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, it.Teardown())
	}()

	r := test.NewKnRunResultCollector(t, it)
	defer r.DumpIfFailed()

	tc := sinkprefixTestConfig{}
	assert.NilError(t, tc.setup())
	defer tc.teardown()

	t.Log("Creating a testservice")
	test.ServiceCreate(r, "testsvc0")
	t.Log("create Ping sources with a sink to hello:testsvc0")
	pingSourceCreateWithConfig(r, "testpingsource0", "* * * * */1", "ping", "hello:testsvc0", tc.knConfigPath)

	jpSinkRefNameInSpec := "jsonpath={.spec.sink.ref.name}"
	out, err := test.GetResourceFieldsWithJSONPath(t, it, "pingsource", "testpingsource0", jpSinkRefNameInSpec)
	assert.NilError(t, err)
	assert.Equal(t, out, "testsvc0")

	t.Log("delete Ping sources")
	pingSourceDelete(r, "testpingsource0")
}

func pingSourceCreateWithConfig(r *test.KnRunResultCollector, sourceName string, schedule string, data string, sink string, config string) {
	out := r.KnTest().Kn().Run("source", "ping", "create", sourceName,
		"--schedule", schedule, "--data", data, "--sink", sink, "--config", config)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "ping", "source", sourceName, "created", "namespace", r.KnTest().Kn().Namespace()))
	r.AssertNoError(out)
}
