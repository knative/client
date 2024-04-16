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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"knative.dev/client-pkg/pkg/util"
)

func TestSecretList(t *testing.T) {
	fakeClient := fake.NewSimpleClientset(
		&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "default"}},
		&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "bar", Namespace: "default"}})

	output, err := executeSecretCommand(fakeClient, "list")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "foo", "bar"))
}

func TestSecretListEmpty(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()

	output, err := executeSecretCommand(fakeClient, "list")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAllIgnoreCase(output, "no", "secret", "found"))
}
