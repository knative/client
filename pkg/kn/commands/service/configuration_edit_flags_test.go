// Copyright © 2022 The Knative Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"testing"

	"gotest.tools/v3/assert"
	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/util"
)

func TestApplyPullPolicyFlag(t *testing.T) {
	var editFlags ConfigurationEditFlags
	knParams := &commands.KnParams{}
	cmd, _, _ := commands.CreateTestKnCommand(NewServiceCreateCommand(knParams), knParams)

	editFlags.AddCreateFlags(cmd)
	svc := createTestService("test-svc", []string{"test-svc-00001", "test-svc-00002"}, goodConditions())
	cmd.SetArgs([]string{"--pull-policy", "Always"})
	cmd.Execute()
	err := editFlags.Apply(&svc, nil, cmd)
	assert.NilError(t, err)
}

func TestApplyPullPolicyFlagError(t *testing.T) {
	var editFlags ConfigurationEditFlags
	knParams := &commands.KnParams{}
	cmd, _, _ := commands.CreateTestKnCommand(NewServiceCreateCommand(knParams), knParams)

	editFlags.AddCreateFlags(cmd)
	svc := createTestService("test-svc", []string{"test-svc-00001", "test-svc-00002"}, goodConditions())
	cmd.SetArgs([]string{"--pull-policy", "InvalidPolicy"})
	cmd.Execute()
	err := editFlags.Apply(&svc, nil, cmd)
	assert.Assert(t, util.ContainsAll(err.Error(), "invalid", "InvalidPolicy", "Valid arguments: Always|Never|IfNotPresent"))
}
