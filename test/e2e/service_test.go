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

//go:build e2e && !eventing
// +build e2e,!eventing

package e2e

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
	v1 "k8s.io/api/core/v1"
	"knative.dev/client/lib/test"
	"knative.dev/client/pkg/util"
	network "knative.dev/networking/pkg/apis/networking"
	pkgtest "knative.dev/pkg/test"
	"knative.dev/serving/pkg/apis/serving"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

const (
	TestPVCSpec = `kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: test-pvc
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
`
)

func TestService(t *testing.T) {
	t.Parallel()
	it, err := test.NewKnTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, it.Teardown())
	}()

	r := test.NewKnRunResultCollector(t, it)
	defer r.DumpIfFailed()

	t.Log("create hello service, delete, and try to create duplicate and get service already exists error")
	test.ServiceCreate(r, "hello")
	serviceCreatePrivate(r, "hello-private")
	serviceCreateDuplicate(r, "hello-private")

	t.Log("return valid info about hello service with print flags")
	serviceDescribeWithPrintFlags(r, "hello")

	t.Log("delete hello service repeatedly and get an error")
	test.ServiceDelete(r, "hello")
	serviceDeleteNonexistent(r, "hello")

	t.Log("delete two services with a service nonexistent")
	test.ServiceCreate(r, "hello")
	serviceMultipleDelete(r, "hello", "bla123")

	t.Log("create service private and make public")
	serviceCreatePrivateUpdatePublic(r, "hello-private-public")

	t.Log("error message from --untag with tag that doesn't exist")
	test.ServiceCreate(r, "untag")
	serviceUntagTagThatDoesNotExist(r, "untag")

	t.Log("delete all services in a namespace")
	test.ServiceCreate(r, "svc1")
	test.ServiceCreate(r, "service2")
	test.ServiceCreate(r, "ksvc3")
	serviceDeleteAll(r)

	t.Log("create services with volume mounts and subpaths")
	serviceCreateWithMount(r)
}

func serviceCreatePrivate(r *test.KnRunResultCollector, serviceName string) {
	out := r.KnTest().Kn().Run("service", "create", serviceName,
		"--image", pkgtest.ImagePath("helloworld"), "--cluster-local")
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "service", serviceName, "creating", "namespace", r.KnTest().Kn().Namespace(), "ready"))

	out = r.KnTest().Kn().Run("service", "describe", serviceName, "--verbose")
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, network.VisibilityLabelKey, serving.VisibilityClusterLocal))
}

func serviceCreatePrivateUpdatePublic(r *test.KnRunResultCollector, serviceName string) {
	out := r.KnTest().Kn().Run("service", "create", serviceName,
		"--image", pkgtest.ImagePath("helloworld"), "--cluster-local")
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "service", serviceName, "creating", "namespace", r.KnTest().Kn().Namespace(), "ready"))

	out = r.KnTest().Kn().Run("service", "describe", serviceName, "--verbose")
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, network.VisibilityLabelKey, serving.VisibilityClusterLocal))

	out = r.KnTest().Kn().Run("service", "update", serviceName, "--no-cluster-local")
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "service", serviceName, "no new revision", "namespace", r.KnTest().Kn().Namespace()))

	out = r.KnTest().Kn().Run("service", "describe", serviceName, "--verbose")
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsNone(out.Stdout, network.VisibilityLabelKey, serving.VisibilityClusterLocal))
}

func serviceCreateDuplicate(r *test.KnRunResultCollector, serviceName string) {
	out := r.KnTest().Kn().Run("service", "list", serviceName)
	r.AssertNoError(out)
	assert.Check(r.T(), strings.Contains(out.Stdout, serviceName), "The service does not exist yet")

	out = r.KnTest().Kn().Run("service", "create", serviceName, "--image", pkgtest.ImagePath("helloworld"))
	r.AssertError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stderr, "the service already exists"))
}

func serviceDescribeWithPrintFlags(r *test.KnRunResultCollector, serviceName string) {
	out := r.KnTest().Kn().Run("service", "describe", serviceName, "-o=name")
	r.AssertNoError(out)

	expectedName := fmt.Sprintf("service.serving.knative.dev/%s", serviceName)
	assert.Equal(r.T(), strings.TrimSpace(out.Stdout), expectedName)
}

func serviceDeleteNonexistent(r *test.KnRunResultCollector, serviceName string) {
	out := r.KnTest().Kn().Run("service", "list", serviceName)
	r.AssertNoError(out)
	assert.Check(r.T(), !strings.Contains(out.Stdout, serviceName), "The service exists")

	out = r.KnTest().Kn().Run("service", "delete", serviceName)
	r.AssertError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stderr, "hello", "not found"), "Failed to get 'not found' error")
}

func serviceMultipleDelete(r *test.KnRunResultCollector, existService, nonexistService string) {
	out := r.KnTest().Kn().Run("service", "list")
	r.AssertNoError(out)
	assert.Check(r.T(), strings.Contains(out.Stdout, existService), "The service ", existService, " does not exist (but is expected to exist)")
	assert.Check(r.T(), !strings.Contains(out.Stdout, nonexistService), "The service", nonexistService, " exists (but is supposed to be not)")

	out = r.KnTest().Kn().Run("service", "delete", existService, nonexistService)
	r.AssertError(out)

	expectedSuccess := fmt.Sprintf(`Service '%s' successfully deleted in namespace '%s'.`, existService, r.KnTest().Kn().Namespace())
	expectedErr := fmt.Sprintf(`services.serving.knative.dev "%s" not found`, nonexistService)
	assert.Check(r.T(), strings.Contains(out.Stdout, expectedSuccess), "Failed to get 'successfully deleted' message")
	assert.Check(r.T(), strings.Contains(out.Stderr, expectedErr), "Failed to get 'not found' error")
}

func serviceUntagTagThatDoesNotExist(r *test.KnRunResultCollector, serviceName string) {
	out := r.KnTest().Kn().Run("ksvc", "ls", serviceName)
	r.AssertNoError(out)
	assert.Check(r.T(), strings.Contains(out.Stdout, serviceName), "Service "+serviceName+" does not exist for test (but should exist)")

	out = r.KnTest().Kn().Run("service", "update", serviceName, "--untag", "foo", "--no-wait")
	assert.Check(r.T(), util.ContainsAll(out.Stderr, "tag(s)", "foo", "not present", "service", "untag"), "Expected error message for using --untag with nonexistent tag")
}

func serviceDeleteAll(r *test.KnRunResultCollector) {
	out := r.KnTest().Kn().Run("services", "ls")
	r.AssertNoError(out)
	// Check if services created successfully/available for test.
	assert.Check(r.T(), !strings.Contains(out.Stdout, "No services found."), "No services created for kn service delete --all e2e (but should exist)")

	out = r.KnTest().Kn().Run("services", "delete", "--all")
	r.AssertNoError(out)
	// Check if output contains successfully deleted to verify deletion took place.
	assert.Check(r.T(), strings.Contains(out.Stdout, "successfully deleted"), "Failed to get 'successfully deleted' message")

	out = r.KnTest().Kn().Run("services", "list")
	r.AssertNoError(out)
	// Check if no services present after kn service delete --all.
	assert.Check(r.T(), strings.Contains(out.Stdout, "No services found."), "Failed to show 'No services found' after kn service delete --all")
}

func serviceCreateWithMount(r *test.KnRunResultCollector) {
	it := r.KnTest()
	kubectl := test.NewKubectl("knative-serving")

	_, err := kubectl.Run("patch", "cm", "config-features", "--patch={\"data\":{\"kubernetes.podspec-persistent-volume-claim\": \"enabled\", "+
		"\"kubernetes.podspec-persistent-volume-write\": \"enabled\", "+
		"\"kubernetes.podspec-volumes-emptydir\": \"enabled\"}}")
	assert.NilError(r.T(), err)
	defer kubectl.Run("patch", "cm", "config-features", "--patch={\"data\":{\"kubernetes.podspec-persistent-volume-claim\": \"disabled\", "+
		"\"kubernetes.podspec-persistent-volume-write\": \"disabled\", "+
		"\"kubernetes.podspec-volumes-emptydir\": \"disabled\"}}")

	kubectl = test.NewKubectl(it.Namespace())

	r.T().Log("create cm test-cm")
	_, err = kubectl.Run("create", "configmap", "test-cm", "--from-literal=key=value")
	assert.NilError(r.T(), err)

	r.T().Log("create service with configmap mounted")
	out := r.KnTest().Kn().Run("service", "create", "test-svc", "--image", pkgtest.ImagePath("helloworld"), "--mount", "/mydir=cm:test-cm")
	r.AssertNoError(out)

	r.T().Log("update the subpath in mounted cm")
	out = r.KnTest().Kn().Run("service", "update", "test-svc", "--mount", "/mydir=cm:test-cm/key")
	r.AssertNoError(out)
	serviceDescribeMount(r, "test-svc", "/mydir", "key")

	r.T().Log("create secret test-sec")
	_, err = kubectl.Run("create", "secret", "generic", "test-sec", "--from-literal", "key1=val1")
	assert.NilError(r.T(), err)

	r.T().Log("update service with a new mount")
	out = r.KnTest().Kn().Run("service", "update", "test-svc", "--mount", "/mydir2=sc:test-sec/key1")
	r.AssertNoError(out)

	r.T().Log("update service with a new emptyDir mount")
	out = r.KnTest().Kn().Run("service", "update", "test-svc", "--mount", "/mydir3=ed:myvol")
	r.AssertNoError(out)

	r.T().Log("update service with a new emptyDir mount with Memory and dir size")
	out = r.KnTest().Kn().Run("service", "update", "test-svc", "--mount", "/mydir4=ed:myvol:type=Memory,size=100Mi")
	r.AssertNoError(out)

	r.T().Log("create PVC test-pvc")
	fp, err := ioutil.TempFile("", "my-pvc")
	assert.NilError(r.T(), err)
	fmt.Fprintf(fp, "%s", TestPVCSpec)
	defer os.Remove(fp.Name())

	_, err = kubectl.Run("create", "-f", fp.Name())
	assert.NilError(r.T(), err)
	r.AssertNoError(out)

	r.T().Log("update service with a new pvc mount")
	out = r.KnTest().Kn().Run("service", "update", "test-svc", "--mount", "/mydir5=pvc:test-pvc")
	r.AssertNoError(out)
}

func getVolumeMountWithHostPath(svc *servingv1.Service, hostPath string) *v1.VolumeMount {
	vols := svc.Spec.Template.Spec.Containers[0].VolumeMounts

	for _, v := range vols {
		if v.MountPath == hostPath {
			return &v
		}
	}
	return nil
}
func serviceDescribeMount(r *test.KnRunResultCollector, serviceName, hostPath, subPath string) {
	out := r.KnTest().Kn().Run("service", "describe", serviceName, "-o=json")
	r.AssertNoError(out)

	svc := &servingv1.Service{}
	assert.NilError(r.T(), json.Unmarshal([]byte(out.Stdout), &svc))

	r.T().Log("check volume mounts not nil")
	volumeMount := getVolumeMountWithHostPath(svc, hostPath)
	assert.Check(r.T(), volumeMount != nil)

	r.T().Log("check volume mount subpath is the same as given")
	assert.Equal(r.T(), subPath, volumeMount.SubPath)
}
