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

//go:build e2e && !serving
// +build e2e,!serving

package e2e

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"knative.dev/client-pkg/pkg/util"
	pkgtest "knative.dev/pkg/test"

	"gotest.tools/v3/assert"

	"knative.dev/client-pkg/pkg/util/test"
)

func TestMultiContainers(t *testing.T) {
	t.Parallel()
	it, err := test.NewKnTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, it.Teardown())
	}()

	r := test.NewKnRunResultCollector(t, it)
	defer r.DumpIfFailed()

	tempDir := t.TempDir()

	t.Log("Creating a multicontainer service from file")
	container := addContainer(r, "sidecar", pkgtest.ImagePath("sidecarcontainer"))
	test.CreateFile("sidecar.yaml", container.Stdout, tempDir, test.FileModeReadWrite)
	createServiceWithSidecar(r, "testsvc0", filepath.Join(tempDir, "sidecar.yaml"))
	test.ServiceDelete(r, "testsvc0")

	t.Log("Creating a multicontainer service from os.Stdin")
	container = addContainer(r, "sidecar", pkgtest.ImagePath("sidecarcontainer"))
	out := createServiceWithPipeInput(r, "testsvc1", container.Stdout)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "service", "testsvc1", "creating", "namespace", r.KnTest().Kn().Namespace(), "ready"))
	test.ServiceDelete(r, "testsvc1")

	t.Log("Creating a multicontainer service from os.Stdin with EOF error")
	out = createServiceWithPipeInput(r, "testsvc2", "")
	r.AssertError(out)
	assert.Assert(t, util.ContainsAllIgnoreCase(out.Stderr, "error", "EOF"))

	t.Log("Adding container no error")
	out = addContainer(r, "foo", pkgtest.ImagePath("sidecarcontainer"))
	r.AssertNoError(out)
	assert.Assert(t, util.ContainsAllIgnoreCase(out.Stdout, "containers", "foo", pkgtest.ImagePath("sidecarcontainer")))

	t.Log("Adding container without name error")
	out = addContainer(r, "", pkgtest.ImagePath("sidecarcontainer"))
	r.AssertError(out)
	assert.Assert(t, util.ContainsAllIgnoreCase(out.Stderr, "requires", "container", "name"))

	t.Log("Adding container without image error")
	out = addContainer(r, "foo", "")
	r.AssertError(out)
	assert.Assert(t, util.ContainsAllIgnoreCase(out.Stderr, "requires", "image", "name"))
}

func addContainer(r *test.KnRunResultCollector, containerName, image string) test.KnRunResult {
	args := []string{"container", "add"}
	if containerName != "" {
		args = append(args, containerName)
	}
	if image != "" {
		args = append(args, "--image", image)
	}

	out := r.KnTest().Kn().RunNoNamespace(args...)
	return out
}

func createServiceWithSidecar(r *test.KnRunResultCollector, serviceName, file string) {
	args := []string{"service", "create", serviceName,
		"--image", pkgtest.ImagePath("servingcontainer"),
		"--port", "8881",
		"--containers", file,
	}

	out := r.KnTest().Kn().Run(args...)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "service", serviceName, "creating", "namespace", r.KnTest().Kn().Namespace(), "ready"))
}

func createServiceWithPipeInput(r *test.KnRunResultCollector, serviceName, sideCar string) test.KnRunResult {
	var stderr bytes.Buffer
	var stdout bytes.Buffer

	args := []string{"service", "create", serviceName,
		"--image", pkgtest.ImagePath("servingcontainer"),
		"--port", "8881",
		"--containers", "-",
		"--namespace", r.KnTest().Kn().Namespace(),
	}

	cmd := exec.Command("kn", args...)

	stdin, err := cmd.StdinPipe()
	assert.NilError(r.T(), err)
	go func() {
		defer stdin.Close()
		io.WriteString(stdin, sideCar)
	}()
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	err = cmd.Run()
	result := test.KnRunResult{
		CmdLine: fmt.Sprintf("%s %s", "kn", strings.Join(args, " ")),
		Stdout:  stdout.String(),
		Stderr:  stderr.String(),
		Error:   err,
	}
	return result
}
