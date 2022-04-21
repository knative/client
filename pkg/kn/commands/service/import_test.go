// Copyright © 2020 The Knative Authors
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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	knclient "knative.dev/client/pkg/serving/v1"
	"knative.dev/client/pkg/util"
	"knative.dev/client/pkg/util/mock"
	"knative.dev/client/pkg/wait"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func TestServiceImportFilenameError(t *testing.T) {
	client := knclient.NewMockKnServiceClient(t)
	r := client.Recorder()

	_, err := executeServiceCommand(client, "import")

	assert.Assert(t, err != nil)
	assert.Assert(t, util.ContainsAll(err.Error(), "'kn service import'", "requires", "file", "single", "argument"))
	assert.Error(t, err, "'kn service import' requires filename of import file as single argument")
	r.Validate()
}

func TestServiceImportExistError(t *testing.T) {
	file, err := generateFile(t, []byte(exportYAML))
	assert.NilError(t, err)

	client := knclient.NewMockKnServiceClient(t)
	r := client.Recorder()

	r.GetService("foo", nil, nil)
	_, err = executeServiceCommand(client, "import", file)

	assert.Assert(t, err != nil)
	assert.Assert(t, util.ContainsAll(err.Error(), "'foo'", "default", "service", "already", "exists"))
	r.Validate()
}

func TestServiceImport(t *testing.T) {
	file, err := generateFile(t, []byte(exportYAML))
	assert.NilError(t, err)

	client := knclient.NewMockKnServiceClient(t)
	r := client.Recorder()

	r.GetService("foo", nil, errors.NewNotFound(servingv1.Resource("service"), "foo"))
	r.CreateService(mock.Any(), nil)
	r.GetConfiguration("foo", nil, nil)
	r.WaitForService("foo", mock.Any(), wait.NoopMessageCallback(), nil, time.Second)
	r.GetService("foo", getServiceWithUrl("foo", "http://foo.example.com"), nil)

	out, err := executeServiceCommand(client, "import", file)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(out, "Service", "'foo'", "default", "imported"))
	r.Validate()
}

func TestServiceImportNoWait(t *testing.T) {
	file, err := generateFile(t, []byte(exportYAML))
	assert.NilError(t, err)

	client := knclient.NewMockKnServiceClient(t)
	r := client.Recorder()

	r.GetService("foo", nil, errors.NewNotFound(servingv1.Resource("service"), "foo"))
	r.CreateService(mock.Any(), nil)
	r.GetConfiguration("foo", nil, nil)

	out, err := executeServiceCommand(client, "import", file, "--no-wait")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(out, "Service", "imported", "foo"))
	r.Validate()
}

func TestServiceImportWitRevisions(t *testing.T) {
	file, err := generateFile(t, []byte(exportWithRevisionsYAML))
	assert.NilError(t, err)

	client := knclient.NewMockKnServiceClient(t)
	r := client.Recorder()

	r.GetService("foo", nil, errors.NewNotFound(servingv1.Resource("service"), "foo"))
	r.CreateService(mock.Any(), nil)
	r.GetConfiguration("foo", getConfiguration("foo"), nil)
	// 2 previous Revisions to re-create + 1 latest from CreateService
	r.CreateRevision(mock.Any(), nil)
	r.CreateRevision(mock.Any(), nil)
	r.WaitForService("foo", mock.Any(), wait.NoopMessageCallback(), nil, time.Second)
	r.GetService("foo", getServiceWithUrl("foo", "http://foo.example.com"), nil)

	out, err := executeServiceCommand(client, "import", file)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(out, "Service", "imported", "foo"))
	r.Validate()
}

func generateFile(t *testing.T, fileContent []byte) (string, error) {
	tempDir := t.TempDir()

	tempFile := filepath.Join(tempDir, "import.yaml")
	if err := ioutil.WriteFile(tempFile, fileContent, os.FileMode(0666)); err != nil {
		return "", err
	}
	return tempFile, nil
}

func getConfiguration(name string) *servingv1.Configuration {
	return &servingv1.Configuration{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

var exportYAML = `
apiVersion: client.knative.dev/v1alpha1
kind: Export
metadata:
  creationTimestamp: null
spec:
  revisions: null
  service:
    apiVersion: serving.knative.dev/v1
    kind: Service
    metadata:
      name: foo
    spec:
      template:
        metadata:
        spec:
          containerConcurrency: 0
          containers:
          - env:
            - name: TARGET
              value: v1
            image: gcr.io/foo/bar:baz
            name: user-container
            readinessProbe:
              successThreshold: 1
              tcpSocket:
                port: 0
            resources: {}
          enableServiceLinks: false
          timeoutSeconds: 300
    status: {}
`

var exportWithRevisionsYAML = `
apiVersion: client.knative.dev/v1alpha1
kind: Export
metadata:
  creationTimestamp: null
spec:
  revisions:
  - apiVersion: serving.knative.dev/v1
    kind: Revision
    metadata:
      annotations:
        client.knative.dev/user-image: gcr.io/foo/bar:baz
        serving.knative.dev/routes: foo
      creationTimestamp: null
      labels:
        serving.knative.dev/configuration: foo
        serving.knative.dev/configurationGeneration: "1"
        serving.knative.dev/routingState: active
        serving.knative.dev/service: foo
      name: foo-rev-1
    spec:
      containerConcurrency: 0
      containers:
      - env:
        - name: TARGET
          value: v1
        image: gcr.io/foo/bar:baz
        name: user-container
        readinessProbe:
          successThreshold: 1
          tcpSocket:
            port: 0
        resources: {}
      enableServiceLinks: false
      timeoutSeconds: 300
    status: {}
  - apiVersion: serving.knative.dev/v1
    kind: Revision
    metadata:
      annotations:
        client.knative.dev/user-image: gcr.io/foo/bar:baz
        serving.knative.dev/routes: foo
      creationTimestamp: null
      labels:
        serving.knative.dev/configuration: foo
        serving.knative.dev/configurationGeneration: "2"
        serving.knative.dev/routingState: active
        serving.knative.dev/service: foo
      name: foo-rev-2
    spec:
      containerConcurrency: 0
      containers:
      - env:
        - name: TARGET
          value: v2
        image: gcr.io/foo/bar:baz
        name: user-container
        readinessProbe:
          successThreshold: 1
          tcpSocket:
            port: 0
        resources: {}
      enableServiceLinks: false
      timeoutSeconds: 300
    status: {}
  service:
    apiVersion: serving.knative.dev/v1
    kind: Service
    metadata:
      creationTimestamp: null
      name: foo
    spec:
      template:
        metadata:
          annotations:
            client.knative.dev/user-image: gcr.io/foo/bar:baz
          name: foo-rev-3
        spec:
          containerConcurrency: 0
          containers:
          - env:
            - name: TARGET
              value: v3
            image: gcr.io/foo/bar:baz
            name: user-container
            readinessProbe:
              successThreshold: 1
              tcpSocket:
                port: 0
            resources: {}
          enableServiceLinks: false
          timeoutSeconds: 300
      traffic:
      - latestRevision: false
        percent: 25
        revisionName: foo-rev-1
      - latestRevision: false
        percent: 25
        revisionName: foo-rev-2
      - latestRevision: false
        percent: 50
        revisionName: foo-rev-3
    status: {}
`
