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
	"strings"
	"testing"
	"time"

	"gotest.tools/assert"
	"k8s.io/apimachinery/pkg/util/wait"
	"knative.dev/client/pkg/util"
)

const (
	// Interval specifies the time between two polls.
	Interval = 10 * time.Second
	// Timeout specifies the timeout for the function PollImmediate to reach a certain status.
	Timeout = 5 * time.Minute
)

func TestTektonPipeline(t *testing.T) {
	test, err := NewE2eTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, test.Teardown())
	}()

	kubectl := kubectl{test.namespace}
	basedir := currentDir(t) + "/../resources/tekton"

	// create secret for the kn-deployer-account service account
	_, err = kubectl.Run("create", "-n", test.namespace, "secret",
		"generic", "container-registry",
		"--from-file=.dockerconfigjson="+Flags.DockerConfigJSON,
		"--type=kubernetes.io/dockerconfigjson")
	assert.NilError(t, err)

	_, err = kubectl.Run("apply", "-f", basedir+"/kn-deployer-rbac.yaml")
	assert.NilError(t, err)

	_, err = kubectl.Run("apply", "-f", basedir+"/buildah.yaml")
	assert.NilError(t, err)

	_, err = kubectl.Run("apply", "-f", "https://raw.githubusercontent.com/tektoncd/catalog/master/kn/kn.yaml")
	assert.NilError(t, err)

	_, err = kubectl.Run("apply", "-f", basedir+"/kn-pipeline.yaml")
	assert.NilError(t, err)

	_, err = kubectl.Run("apply", "-f", basedir+"/kn-pipeline-resource.yaml")
	assert.NilError(t, err)

	_, err = kubectl.Run("create", "-f", basedir+"/kn-pipeline-run.yaml")
	assert.NilError(t, err)

	err = waitForPipelineSuccess(kubectl)
	assert.NilError(t, err)

	r := NewKnRunResultCollector(t)

	const serviceName = "hello"
	out := test.kn.Run("service", "describe", serviceName)
	r.AssertNoError(out)
	assert.Assert(t, util.ContainsAll(out.Stdout, serviceName, test.kn.namespace))
	assert.Assert(t, util.ContainsAll(out.Stdout, "Conditions", "ConfigurationsReady", "Ready", "RoutesReady"))
}

func waitForPipelineSuccess(k kubectl) error {
	return wait.PollImmediate(Interval, Timeout, func() (bool, error) {
		out, err := k.Run("get", "pipelinerun", "-o=jsonpath='{.items[0].status.conditions[?(@.type==\"Succeeded\")].status}'")
		return strings.Contains(out, "True"), err
	})
}
