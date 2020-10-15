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
// +build !eventing

package e2e

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"gotest.tools/assert"
	"knative.dev/client/lib/test"
	"knative.dev/client/pkg/util"
)

func TestServiceImport(t *testing.T) {
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

	t.Log("import service foo with revision")
	testFile := filepath.Join(tempDir, "foo-with-revisions")

	serviceCreateWithOptions(r, "foo", "--revision-name", "foo-rev-1")
	test.ServiceUpdate(r, "foo", "--env", "TARGET=v2", "--revision-name", "foo-rev-2")
	test.ServiceUpdate(r, "foo", "--traffic", "foo-rev-1=50,foo-rev-2=50")
	serviceExportToFile(r, "foo", testFile, true)
	test.ServiceDelete(r, "foo")
	serviceImport(r, testFile)

	t.Log("import existing service foo error")
	serviceImport(r, testFile)

	t.Log("import service from missing file error")
	serviceImportFileError(r, testFile+"-missing")
}

func serviceExportToFile(r *test.KnRunResultCollector, serviceName, filename string, withRevisions bool) {
	command := []string{"service", "export", serviceName, "-o", "yaml", "--mode", "export"}
	if withRevisions {
		command = append(command, "--with-revisions")
	}
	out := r.KnTest().Kn().Run(command...)
	r.AssertNoError(out)
	err := ioutil.WriteFile(filename, []byte(out.Stdout), test.FileModeReadWrite)
	assert.NilError(r.T(), err)
}

func serviceImport(r *test.KnRunResultCollector, filename string) {
	command := []string{"service", "import", filename}

	out := r.KnTest().Kn().Run(command...)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "service", "importing", "namespace", r.KnTest().Kn().Namespace(), "ready"))
}

func serviceImportExistsError(r *test.KnRunResultCollector, filename string) {
	command := []string{"service", "import", filename}

	out := r.KnTest().Kn().Run(command...)
	r.AssertError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stderr, "service", "already", "exists"))
}

func serviceImportFileError(r *test.KnRunResultCollector, filePath string) {
	out := r.KnTest().Kn().Run("service", "import", filePath)
	r.AssertError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stderr, "no", "such", "file", "directory", filePath))
}
