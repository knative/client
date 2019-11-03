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
	"testing"
	"time"

	"gotest.tools/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"

	"knative.dev/serving/pkg/apis/serving/v1alpha1"

	servinglib "knative.dev/client/pkg/serving"
	knclient "knative.dev/client/pkg/serving/v1alpha1"
	"knative.dev/client/pkg/wait"

	"knative.dev/client/pkg/util"
)

func TestServiceCreateImageMock(t *testing.T) {

	// New mock client
	client := knclient.NewMockKnClient(t)

	// Recording:
	r := client.Recorder()
	// Check for existing service --> no
	r.GetService("foo", nil, errors.NewNotFound(v1alpha1.Resource("service"), "foo"))
	// Create service (don't validate given service --> "Any()" arg is allowed)
	r.CreateService(knclient.Any(), nil)
	// Wait for service to become ready
	r.WaitForService("foo", knclient.Any(), wait.NoopMessageCallback(), nil, time.Second)
	// Get for showing the URL
	r.GetService("foo", getServiceWithUrl("foo", "http://foo.example.com"), nil)

	// Testing:
	output, err := executeServiceCommand(client, "create", "foo", "--image", "gcr.io/foo/bar:baz")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "Creating", "foo", "http://foo.example.com", "Ready"))

	// Validate that all recorded API methods have been called
	r.Validate()
}

func TestServiceCreateEnvMock(t *testing.T) {
	client := knclient.NewMockKnClient(t)

	r := client.Recorder()
	r.GetService("foo", nil, errors.NewNotFound(v1alpha1.Resource("service"), "foo"))

	service := getService("foo")
	envVars := []corev1.EnvVar{
		{Name: "a", Value: "mouse"},
		{Name: "b", Value: "cookie"},
		{Name: "empty", Value: ""},
	}
	template, err := servinglib.RevisionTemplateOfService(service)
	assert.NilError(t, err)
	template.Spec.GetContainer().Env = envVars
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{servinglib.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}
	r.CreateService(service, nil)

	output, err := executeServiceCommand(client, "create", "foo", "--image", "gcr.io/foo/bar:baz", "-e", "a=mouse", "--env", "b=cookie", "--env=empty", "--async", "--revision-name=")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "created", "foo", "default"))

	r.Validate()
}

func TestServiceCreateLabel(t *testing.T) {
	client := knclient.NewMockKnClient(t)

	r := client.Recorder()
	r.GetService("foo", nil, errors.NewNotFound(v1alpha1.Resource("service"), "foo"))

	service := getService("foo")
	expected := map[string]string{
		"a":     "mouse",
		"b":     "cookie",
		"empty": "",
	}
	service.Labels = expected
	service.Spec.Template.Annotations = map[string]string{
		servinglib.UserImageAnnotationKey: "gcr.io/foo/bar:baz",
	}
	template, err := servinglib.RevisionTemplateOfService(service)
	assert.NilError(t, err)
	template.ObjectMeta.Labels = expected
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	r.CreateService(service, nil)

	output, err := executeServiceCommand(client, "create", "foo", "--image", "gcr.io/foo/bar:baz", "-l", "a=mouse", "--label", "b=cookie", "--label=empty", "--async", "--revision-name=")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "created", "foo", "default"))

	r.Validate()
}

func TestServiceCreateWithEnvFromConfigMap(t *testing.T) {
	client := knclient.NewMockKnClient(t)

	r := client.Recorder()
	r.GetService("foo", nil, errors.NewNotFound(v1alpha1.Resource("service"), "foo"))

	service := getService("foo")
	template, err := servinglib.RevisionTemplateOfService(service)
	assert.NilError(t, err)
	template.Spec.GetContainer().EnvFrom = []corev1.EnvFromSource{
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "config-map-name",
				},
			},
		},
	}
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{servinglib.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}
	r.CreateService(service, nil)

	output, err := executeServiceCommand(client, "create", "foo", "--image", "gcr.io/foo/bar:baz", "--env-from", "config-map:config-map-name", "--async", "--revision-name=")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "created", "foo", "default"))

	r.Validate()
}

func TestServiceCreateWithEnvFromConfigMapRemoval(t *testing.T) {
	client := knclient.NewMockKnClient(t)

	r := client.Recorder()
	r.GetService("foo", nil, errors.NewNotFound(v1alpha1.Resource("service"), "foo"))

	service := getService("foo")
	template, err := servinglib.RevisionTemplateOfService(service)
	assert.NilError(t, err)
	template.Spec.GetContainer().EnvFrom = nil
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{servinglib.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}
	r.CreateService(service, nil)

	output, err := executeServiceCommand(client, "create", "foo", "--image", "gcr.io/foo/bar:baz", "--env-from", "config-map:config-map-name-", "--async", "--revision-name=")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "created", "foo", "default"))

	r.Validate()
}

func TestServiceCreateWithEnvFromEmptyRemoval(t *testing.T) {
	client := knclient.NewMockKnClient(t)

	r := client.Recorder()
	r.GetService("foo", nil, errors.NewNotFound(v1alpha1.Resource("service"), "foo"))

	service := getService("foo")
	template, err := servinglib.RevisionTemplateOfService(service)
	assert.NilError(t, err)
	template.Spec.GetContainer().EnvFrom = nil
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{servinglib.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}
	r.CreateService(service, nil)

	_, err = executeServiceCommand(client, "create", "foo", "--image", "gcr.io/foo/bar:baz", "--env-from", "-", "--async", "--revision-name=")
	assert.Error(t, err, "\"-\" is not a valid value for \"--env-from\"")
}

func TestServiceCreateWithEnvFromSecret(t *testing.T) {
	client := knclient.NewMockKnClient(t)

	r := client.Recorder()
	r.GetService("foo", nil, errors.NewNotFound(v1alpha1.Resource("service"), "foo"))

	service := getService("foo")
	template, err := servinglib.RevisionTemplateOfService(service)
	assert.NilError(t, err)
	template.Spec.GetContainer().EnvFrom = []corev1.EnvFromSource{
		{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "secret-name",
				},
			},
		},
	}
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{servinglib.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}
	r.CreateService(service, nil)

	output, err := executeServiceCommand(client, "create", "foo", "--image", "gcr.io/foo/bar:baz", "--env-from", "secret:secret-name", "--async", "--revision-name=")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "created", "foo", "default"))

	r.Validate()
}

func TestServiceCreateWithEnvFromSecretRemoval(t *testing.T) {
	client := knclient.NewMockKnClient(t)

	r := client.Recorder()
	r.GetService("foo", nil, errors.NewNotFound(v1alpha1.Resource("service"), "foo"))

	service := getService("foo")
	template, err := servinglib.RevisionTemplateOfService(service)
	assert.NilError(t, err)
	template.Spec.GetContainer().EnvFrom = nil
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{servinglib.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}
	r.CreateService(service, nil)

	output, err := executeServiceCommand(client, "create", "foo", "--image", "gcr.io/foo/bar:baz", "--env-from", "secret:secret-name-", "--async", "--revision-name=")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "created", "foo", "default"))

	r.Validate()
}

func TestServiceCreateWithVolumeAndMountConfigMap(t *testing.T) {
	client := knclient.NewMockKnClient(t)

	r := client.Recorder()
	r.GetService("foo", nil, errors.NewNotFound(v1alpha1.Resource("service"), "foo"))

	service := getService("foo")
	template, err := servinglib.RevisionTemplateOfService(service)
	assert.NilError(t, err)
	template.Spec.Volumes = []corev1.Volume{
		{
			Name: "volume-name",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "config-map-name",
					},
				},
			},
		},
	}

	template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
		{
			Name:      "volume-name",
			MountPath: "/mount/path",
			ReadOnly:  true,
		},
	}

	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{servinglib.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}
	r.CreateService(service, nil)

	output, err := executeServiceCommand(client, "create", "foo", "--image", "gcr.io/foo/bar:baz",
		"--mount", "/mount/path=volume-name", "--volume", "volume-name=cm:config-map-name", "--async", "--revision-name=")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "created", "foo", "default"))

	r.Validate()
}

func TestServiceCreateWithMountConfigMap(t *testing.T) {
	client := knclient.NewMockKnClient(t)

	r := client.Recorder()
	r.GetService("foo", nil, errors.NewNotFound(v1alpha1.Resource("service"), "foo"))

	service := getService("foo")
	template, err := servinglib.RevisionTemplateOfService(service)
	assert.NilError(t, err)
	template.Spec.Volumes = []corev1.Volume{
		{
			Name: "kn-managed-volume:/mount/path",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "config-map-name",
					},
				},
			},
		},
	}

	template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
		{
			Name:      "kn-managed-volume:/mount/path",
			MountPath: "/mount/path",
			ReadOnly:  true,
		},
	}

	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{servinglib.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}
	r.CreateService(service, nil)

	output, err := executeServiceCommand(client, "create", "foo", "--image", "gcr.io/foo/bar:baz",
		"--mount", "/mount/path=cm:config-map-name", "--async", "--revision-name=")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "created", "foo", "default"))

	r.Validate()
}

func TestServiceCreateWithVolumeAndMountSecret(t *testing.T) {
	client := knclient.NewMockKnClient(t)

	r := client.Recorder()
	r.GetService("foo", nil, errors.NewNotFound(v1alpha1.Resource("service"), "foo"))

	service := getService("foo")
	template, err := servinglib.RevisionTemplateOfService(service)
	assert.NilError(t, err)
	template.Spec.Volumes = []corev1.Volume{
		{
			Name: "volume-name",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: "secret-name",
				},
			},
		},
	}

	template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
		{
			Name:      "volume-name",
			MountPath: "/mount/path",
			ReadOnly:  true,
		},
	}

	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{servinglib.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}
	r.CreateService(service, nil)

	output, err := executeServiceCommand(client, "create", "foo", "--image", "gcr.io/foo/bar:baz",
		"--mount", "/mount/path=volume-name", "--volume", "volume-name=secret:secret-name", "--async", "--revision-name=")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "created", "foo", "default"))

	r.Validate()
}

func TestServiceCreateWithMountSecret(t *testing.T) {
	client := knclient.NewMockKnClient(t)

	r := client.Recorder()
	r.GetService("foo", nil, errors.NewNotFound(v1alpha1.Resource("service"), "foo"))

	service := getService("foo")
	template, err := servinglib.RevisionTemplateOfService(service)
	assert.NilError(t, err)
	template.Spec.Volumes = []corev1.Volume{
		{
			Name: "kn-managed-volume:/mount/path",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: "secret-name",
				},
			},
		},
	}

	template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
		{
			Name:      "kn-managed-volume:/mount/path",
			MountPath: "/mount/path",
			ReadOnly:  true,
		},
	}

	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{servinglib.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}
	r.CreateService(service, nil)

	output, err := executeServiceCommand(client, "create", "foo", "--image", "gcr.io/foo/bar:baz",
		"--mount", "/mount/path=sc:secret-name", "--async", "--revision-name=")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "created", "foo", "default"))

	r.Validate()
}

func getService(name string) *v1alpha1.Service {
	service := &v1alpha1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: v1alpha1.ServiceSpec{},
	}

	service.Spec.Template = &v1alpha1.RevisionTemplateSpec{
		Spec: v1alpha1.RevisionSpec{},
	}

	service.Spec.Template.Spec.Containers = []corev1.Container{{
		Resources: corev1.ResourceRequirements{
			Limits:   corev1.ResourceList{},
			Requests: corev1.ResourceList{},
		},
	}}
	return service
}

func getServiceWithUrl(name string, urlName string) *v1alpha1.Service {
	service := v1alpha1.Service{}
	url, _ := apis.ParseURL(urlName)
	service.Status.URL = url
	service.Name = name
	return &service
}
