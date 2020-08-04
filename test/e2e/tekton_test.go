// Copyright 2019 The Knative Authors

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build tekton

package e2e

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"gotest.tools/assert"
	"k8s.io/apimachinery/pkg/util/wait"

	"knative.dev/client/lib/test"
	"knative.dev/client/pkg/util"
)

const (
	// Interval specifies the time between two polls.
	Interval = 10 * time.Second
	// Timeout specifies the timeout for the function PollImmediate to reach a certain status.
	Timeout = 5 * time.Minute
)

func TestTektonPipeline(t *testing.T) {
	it, err := test.NewKnTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, it.Teardown())
	}()

	kubectl := test.NewKubectl(it.Namespace())
	basedir := test.CurrentDir(t) + "/../resources/tekton"

	// create secret for the kn-deployer-account service account
	_, err = kubectl.Run("create", "-n", it.Namespace(), "secret",
		"generic", "container-registry",
		"--from-file=.dockerconfigjson="+test.Flags.DockerConfigJSON,
		"--type=kubernetes.io/dockerconfigjson")
	assert.NilError(t, err)

	_, err = kubectl.Run("apply", "-f", basedir+"/kn-deployer-rbac.yaml")
	assert.NilError(t, err)

	_, err = kubectl.Run("apply", "-f", tektonCatalogTask("git-clone", "0.1"))
	assert.NilError(t, err)

	_, err = kubectl.Run("apply", "-f", basedir+"/resources.yaml")
	assert.NilError(t, err)

	_, err = kubectl.Run("apply", "-f", tektonCatalogTask("buildah", "0.1"))
	assert.NilError(t, err)

	_, err = kubectl.Run("apply", "-f", tektonCatalogTask("kn", "0.1"))
	assert.NilError(t, err)

	_, err = kubectl.Run("apply", "-f", basedir+"/kn-pipeline.yaml")
	assert.NilError(t, err)

	_, err = kubectl.Run("create", "-f", basedir+"/kn-pipeline-run.yaml")
	assert.NilError(t, err)

	err = waitForPipelineSuccess(kubectl)
	assert.NilError(t, err)

	r := test.NewKnRunResultCollector(t, it)

	const serviceName = "hello"
	out := it.Kn().Run("service", "describe", serviceName)
	r.AssertNoError(out)
	assert.Assert(t, util.ContainsAll(out.Stdout, serviceName, it.Kn().Namespace()))
	assert.Assert(t, util.ContainsAll(out.Stdout, "Conditions", "ConfigurationsReady", "Ready", "RoutesReady"))
}

func waitForPipelineSuccess(k test.Kubectl) error {
	return wait.PollImmediate(Interval, Timeout, func() (bool, error) {
		out, err := k.Run("get", "pipelinerun", "-o=jsonpath='{.items[0].status.conditions[?(@.type==\"Succeeded\")].status}'")
		return strings.Contains(out, "True"), err
	})
}

func tektonCatalogTask(taskName, version string) string {
	return fmt.Sprintf(
		"https://raw.githubusercontent.com/tektoncd/catalog/"+
			"master/task/%s/%s/%s.yaml",
		taskName, version, taskName,
	)
}
