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

package e2e

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestRevisionWorkflow(t *testing.T) {
	teardown := Setup(t)
	defer teardown(t)

	testServiceCreate(t, k, "hello")
	// TODO: remove this when https://github.com/knative/client/pull/156 is merged
	time.Sleep(10 * time.Second)
	testDeleteRevision(t, k, "hello")
	testServiceDelete(t, k, "hello")
}

func testDeleteRevision(t *testing.T, k kn, serviceName string) {
	revName, err := k.RunWithOpts([]string{"revision", "list", "-o=jsonpath={.items[0].metadata.name}"}, runOpts{})
	if err != nil {
		t.Errorf("Error executing 'revision list -o' command. Error: %s", err.Error())
	}
	if strings.Contains(revName, "No resources found.") {
		t.Errorf("Could not find revision name.")
	}
	out, err := k.RunWithOpts([]string{"revision", "delete", revName}, runOpts{})
	if err != nil {
		t.Errorf("Error executing 'revision delete %s' command. Error: %s", revName, err.Error())
	}
	expectedOutput := fmt.Sprintf("Revision '%s' successfully deleted in namespace '%s'.\n", revName, defaultKnE2ENamespace)
	if out != expectedOutput {
		t.Errorf("Wrong output from 'revision delete %s' command. Actual output:\n%s\nExpected output:\n%s\n", revName, out, expectedOutput)
	}
}
