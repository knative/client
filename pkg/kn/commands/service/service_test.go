// Copyright © 2019 The Knative Authors
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
	"bytes"

	"github.com/knative/client/pkg/kn/commands"
	knclient "github.com/knative/client/pkg/serving/v1alpha1"
)

// Helper methods

func executeServiceCommand(client knclient.KnClient, args ...string) (string, error) {
	knParams := &commands.KnParams{}

	output := new(bytes.Buffer)
	knParams.Output = output
	knParams.NewClient = func(namespace string) (knclient.KnClient, error) {
		return client, nil
	}
	cmd := NewServiceCommand(knParams)
	cmd.SetArgs(args)
	cmd.SetOutput(output)
	err := cmd.Execute()
	return output.String(), err
}
