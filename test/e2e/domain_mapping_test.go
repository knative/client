// Copyright 2021 The Knative Authors

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build e2e && !eventing
// +build e2e,!eventing

package e2e

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"knative.dev/client/pkg/util"

	"gotest.tools/v3/assert"

	"knative.dev/client/lib/test"
)

func TestDomain(t *testing.T) {
	t.Parallel()
	it, err := test.NewKnTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, it.Teardown())
	}()

	r := test.NewKnRunResultCollector(t, it)
	defer r.DumpIfFailed()

	kubectl := test.NewKubectl("knative-serving")
	_, err = kubectl.Run("patch", "cm", "config-network", "--patch={\"data\":{\"autocreateClusterDomainClaims\":\"true\"}}")
	assert.NilError(t, err)
	defer kubectl.Run("patch", "cm", "config-network", "--patch={\"data\":{\"autocreateClusterDomainClaims\":\"false\"}}")

	domainName := "hello.example.com"

	t.Log("create domain mapping to hello ksvc")
	test.ServiceCreate(r, "hello")
	domainCreate(r, domainName, "hello")

	t.Log("list domain mappings")
	domainList(r, domainName)

	t.Log("update domain mapping Knative service reference")
	test.ServiceCreate(r, "foo")
	domainUpdate(r, domainName, "foo")

	t.Log("describe domain mappings")
	domainDescribe(r, domainName, false)

	t.Log("delete domain")
	domainDelete(r, domainName)

	tempDir := t.TempDir()
	defer os.Remove(tempDir)

	cert, key := makeCertificateForDomain(t, "newdomain.com")

	err = os.WriteFile(filepath.Join(tempDir, "cert"), cert, test.FileModeReadWrite)
	assert.NilError(t, err)
	err = os.WriteFile(filepath.Join(tempDir, "key"), key, test.FileModeReadWrite)
	assert.NilError(t, err)

	kubectl = test.NewKubectl(it.Namespace())
	_, err = kubectl.Run("create", "secret", "tls", "tls-secret", "--cert="+filepath.Join(tempDir, "cert"), "--key="+filepath.Join(tempDir, "key"))
	assert.NilError(t, err)

	t.Log("create domain with TLS")
	domainCreateWithTls(r, "newdomain.com", "hello", "tls-secret")
	time.Sleep(time.Second)
	domainDescribe(r, "newdomain.com", true)
}

func domainCreate(r *test.KnRunResultCollector, domainName, serviceName string, options ...string) {
	command := []string{"domain", "create", domainName, "--ref", serviceName}
	command = append(command, options...)
	out := r.KnTest().Kn().Run(command...)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "domain", "mapping", serviceName, domainName, "created", "namespace", r.KnTest().Kn().Namespace()))
}

func domainCreateWithTls(r *test.KnRunResultCollector, domainName, serviceName, tls string, options ...string) {
	command := []string{"domain", "create", domainName, "--ref", serviceName, "--tls", tls}
	command = append(command, options...)
	out := r.KnTest().Kn().Run(command...)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "domain", "mapping", domainName, "created", "namespace", r.KnTest().Kn().Namespace()))
}

func domainUpdate(r *test.KnRunResultCollector, domainName, serviceName string, options ...string) {
	command := []string{"domain", "update", domainName, "--ref", serviceName}
	command = append(command, options...)
	out := r.KnTest().Kn().Run(command...)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "domain", "mapping", domainName, "updated", "namespace", r.KnTest().Kn().Namespace()))
}

func domainDelete(r *test.KnRunResultCollector, domainName string, options ...string) {
	command := []string{"domain", "delete", domainName}
	command = append(command, options...)
	out := r.KnTest().Kn().Run(command...)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "domain", "mapping", domainName, "deleted", "namespace", r.KnTest().Kn().Namespace()))
}

func domainList(r *test.KnRunResultCollector, domainName string) {
	out := r.KnTest().Kn().Run("domain", "list")
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stdout, domainName))
}

func domainDescribe(r *test.KnRunResultCollector, domainName string, tls bool) {
	k := test.NewKubectl(r.KnTest().Kn().Namespace())
	// Wait for Domain Mapping URL to be populated
	for i := 0; i < 10; i++ {
		out, err := k.Run("get", "domainmapping", domainName, "-o=jsonpath='{.status.url}'")
		assert.NilError(r.T(), err)
		if len(out) > 0 {
			break
		}
		time.Sleep(time.Second)
	}
	out := r.KnTest().Kn().Run("domain", "describe", domainName)
	r.AssertNoError(out)
	var url string
	if tls {
		url = "https://" + domainName
	} else {
		url = "http://" + domainName
	}
	assert.Assert(r.T(), util.ContainsAll(out.Stdout, "Name", "Namespace", "URL", "Service", url))
}

func makeCertificateForDomain(t *testing.T, domainName string) (cert []byte, key []byte) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		t.Fatalf("Failed to generate serial number: %v", err)
	}
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}
	public := &priv.PublicKey
	now := time.Now()
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore: now,
		NotAfter:  now.Add(time.Hour * 12),

		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{domainName},
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, public, priv)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}
	var certOut bytes.Buffer
	pem.Encode(&certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatalf("Failed to marshal private key: %v", err)
	}
	var keyOut bytes.Buffer
	pem.Encode(&keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	return certOut.Bytes(), keyOut.Bytes()
}
