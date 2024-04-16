// Copyright © 2021 The Knative Authors
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

package secret

import (
	"testing"

	"gotest.tools/v3/assert"
	"k8s.io/client-go/kubernetes/fake"
	"knative.dev/client-pkg/pkg/util"
)

func TestSecretCreate(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()

	output, err := executeSecretCommand(fakeClient, "create", "foo", "--from-literal", "user=foo")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "created"))
}

func TestSecretCreateError(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()

	_, err := executeSecretCommand(fakeClient, "create", "--from-literal", "user=foo")
	assert.ErrorContains(t, err, "single argument")
}

func TestSecretCreateCertError(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()

	_, err := executeSecretCommand(fakeClient, "create", "foo", "--tls-cert", "foo.cert")
	assert.Assert(t, err != nil)
	assert.Assert(t, util.ContainsAll(err.Error(), "--tls-cert", "--tls-key", "required"))

	_, err = executeSecretCommand(fakeClient, "create", "foo", "-l", "k=v", "--tls-cert", "foo.cert")
	assert.Assert(t, err != nil)
	assert.Assert(t, util.ContainsAll(err.Error(), "combined", "options"))
}
