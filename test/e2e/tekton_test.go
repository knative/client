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
	test := NewE2eTest(t)
	test.Setup(t)

	kubectl := kubectl{t, Logger{}}
	basedir := currentDir(t) + "/../resources/tekton"

	// create secret for the kn-deployer-account service account
	_, err := kubectl.RunWithOpts([]string{"create", "-n", test.env.Namespace, "secret",
		"generic", "container-registry",
		"--from-file=.dockerconfigjson=" + Flags.DockerConfigJSON,
		"--type=kubernetes.io/dockerconfigjson"}, runOpts{})
	assert.NilError(t, err)

	_, err = kubectl.RunWithOpts([]string{"apply", "-n", test.env.Namespace, "-f", basedir + "/kn-deployer-rbac.yaml"}, runOpts{})
	assert.NilError(t, err)

	_, err = kubectl.RunWithOpts([]string{"apply", "-n", test.env.Namespace, "-f", basedir + "/buildah.yaml"}, runOpts{})
	assert.NilError(t, err)

	_, err = kubectl.RunWithOpts([]string{"apply", "-n", test.env.Namespace, "-f", "https://raw.githubusercontent.com/tektoncd/catalog/master/kn/kn.yaml"}, runOpts{})
	assert.NilError(t, err)

	_, err = kubectl.RunWithOpts([]string{"apply", "-n", test.env.Namespace, "-f", basedir + "/kn-pipeline.yaml"}, runOpts{})
	assert.NilError(t, err)

	_, err = kubectl.RunWithOpts([]string{"apply", "-n", test.env.Namespace, "-f", basedir + "/kn-pipeline-resource.yaml"}, runOpts{})
	assert.NilError(t, err)

	_, err = kubectl.RunWithOpts([]string{"create", "-n", test.env.Namespace, "-f", basedir + "/kn-pipeline-run.yaml"}, runOpts{})
	assert.NilError(t, err)

	err = waitForPipelineSuccess(t, kubectl, test.env.Namespace)
	assert.NilError(t, err)

	const serviceName = "hello"
	out, err := test.kn.RunWithOpts([]string{"service", "describe", serviceName}, runOpts{NoNamespace: false})
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(out, serviceName, test.kn.namespace))
	assert.Assert(t, util.ContainsAll(out, "Conditions", "ConfigurationsReady", "Ready", "RoutesReady"))

	// tear down only if the test passes, we want to keep the pods otherwise
	test.Teardown(t)
}

func waitForPipelineSuccess(t *testing.T, k kubectl, namespace string) error {
	return wait.PollImmediate(Interval, Timeout, func() (bool, error) {
		out, err := k.RunWithOpts([]string{"get", "pipelinerun", "-n", namespace, "-o=jsonpath='{.items[0].status.conditions[?(@.type==\"Succeeded\")].status}'"}, runOpts{})
		return strings.Contains(out, "True"), err
	})
}
