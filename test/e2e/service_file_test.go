// Copyright 2020 The Knative Authors

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build e2e && !eventing
// +build e2e,!eventing

package e2e

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"gotest.tools/v3/assert/cmp"

	"gotest.tools/v3/assert"

	"knative.dev/client/lib/test"
	"knative.dev/client/pkg/util"
	pkgtest "knative.dev/pkg/test"
)

const (
	ServiceYAML string = `
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: foo-yaml
spec:
  template:
    spec:
      containers:
        - image: %s
          env:
            - name: TARGET
              value: "Go Sample v1"`

	ServiceJSON string = `
{
  "apiVersion": "serving.knative.dev/v1",
  "kind": "Service",
  "metadata": {
    "name": "foo-json"
  },
  "spec": {
    "template": {
		"spec": {
		"containers": [
			{
			"image": "%s",
			"env": [
				{
				"name": "TARGET",
				"value": "Go Sample v1"
				}
			]
		  }
		]
	  }
	}
  }
}`
)

func TestServiceCreateFromFile(t *testing.T) {
	t.Parallel()
	it, err := test.NewKnTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, it.Teardown())
	}()

	r := test.NewKnRunResultCollector(t, it)
	defer r.DumpIfFailed()

	tempDir, err := ioutil.TempDir("", "kn-file")
	defer os.RemoveAll(tempDir)
	assert.NilError(t, err)

	test.CreateFile("foo.json", fmt.Sprintf(ServiceJSON, pkgtest.ImagePath("helloworld")), tempDir, test.FileModeReadWrite)
	test.CreateFile("foo.yaml", fmt.Sprintf(ServiceYAML, pkgtest.ImagePath("helloworld")), tempDir, test.FileModeReadWrite)

	t.Log("create foo-json service from JSON file")
	serviceCreateFromFile(r, "foo-json", filepath.Join(tempDir, "foo.json"), true)

	t.Log("create foo-yaml service from YAML file")
	serviceCreateFromFile(r, "foo-yaml", filepath.Join(tempDir, "foo.yaml"), false)

	t.Log("error message for non-existing file")
	serviceCreateFromFileError(r, "foo", filepath.Join(tempDir, "fake-foo.json"))
	serviceCreateFromFileError(r, "foo", filepath.Join(tempDir, "fake-foo.yaml"))

	t.Log("error message for mismatch names")
	serviceCreateFromFileNameMismatch(r, "foo", filepath.Join(tempDir, "foo.json"))
	serviceCreateFromFileNameMismatch(r, "foo", filepath.Join(tempDir, "foo.yaml"))
}

func serviceCreateFromFile(r *test.KnRunResultCollector, serviceName, filePath string, useName bool) {
	var out test.KnRunResult
	if useName {
		out = r.KnTest().Kn().Run("service", "create", serviceName, "-f", filePath)
	} else {
		out = r.KnTest().Kn().Run("service", "create", "-f", filePath)
	}
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "service", serviceName, "creating", "namespace", r.KnTest().Kn().Namespace(), "ready"))

	out = r.KnTest().Kn().Run("service", "describe", serviceName, "--verbose")
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, serviceName))
}

func serviceCreateFromFileError(r *test.KnRunResultCollector, serviceName, filePath string) {
	out := r.KnTest().Kn().Run("service", "create", serviceName, "--filename", filePath)
	r.AssertError(out)
	assert.Check(r.T(), fileNotFoundErrorCheck(out, filePath))
}

func serviceCreateFromFileNameMismatch(r *test.KnRunResultCollector, serviceName, filePath string) {
	out := r.KnTest().Kn().Run("service", "create", serviceName, "--filename", filePath)
	r.AssertError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stderr, "provided", "'"+serviceName+"'", "name", "match", "from", "file"))
}

func fileNotFoundErrorCheck(out test.KnRunResult, filePath string) cmp.Comparison {
	result := util.ContainsAllIgnoreCase(out.Stderr, "no", "such", "file", "directory", filePath)
	if result() == cmp.ResultSuccess {
		return result
	}
	// Check for Windows specific error message in case file is not found
	return util.ContainsAllIgnoreCase(out.Stderr, "system", "cannot", "find", "file", "specified", filePath)
}
