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
	"context"
	"os"
	"path/filepath"
	"testing"

	"gotest.tools/v3/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"knative.dev/client/pkg/util"
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

func TestSecretCreateTLS(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()

	dir := t.TempDir()
	certFile := filepath.Join(dir, "tls.crt")
	keyFile := filepath.Join(dir, "tls.key")
	assert.NilError(t, os.WriteFile(certFile, []byte("cert-data"), 0600))
	assert.NilError(t, os.WriteFile(keyFile, []byte("key-data"), 0600))

	output, err := executeSecretCommand(fakeClient, "create", "foo", "--namespace", "default", "--tls-cert", certFile, "--tls-key", keyFile)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "created"))

	secret, err := fakeClient.CoreV1().Secrets("default").Get(context.Background(), "foo", metav1.GetOptions{})
	assert.NilError(t, err)
	assert.Equal(t, secret.Type, corev1.SecretTypeTLS)
	assert.DeepEqual(t, secret.StringData, map[string]string{
		corev1.TLSCertKey:       "cert-data",
		corev1.TLSPrivateKeyKey: "key-data",
	})
}
