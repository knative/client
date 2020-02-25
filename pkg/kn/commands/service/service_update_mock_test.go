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

	"gotest.tools/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	clientserving "knative.dev/client/pkg/serving"
	clientservingv1 "knative.dev/client/pkg/serving/v1"
	"knative.dev/client/pkg/util"
	"knative.dev/pkg/ptr"
)

func TestServiceUpdateEnvMock(t *testing.T) {
	client := clientservingv1.NewMockKnServiceClient(t)

	service := getService("foo")
	template := &service.Spec.Template
	template.Spec.Containers[0].Env = []corev1.EnvVar{
		{Name: "a", Value: "mouse"},
		{Name: "b", Value: "cookie"},
		{Name: "empty", Value: ""},
	}
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	updated := getService("foo")
	template = &updated.Spec.Template
	template.Spec.Containers[0].Env = []corev1.EnvVar{
		{Name: "a", Value: "rabbit"},
		{Name: "b", Value: "cookie"},
	}
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	r := client.Recorder()
	recordServiceUpdateWithSuccess(r, "foo", service, updated)

	output, err := executeServiceCommand(client, "create", "foo", "--image", "gcr.io/foo/bar:baz", "-e", "a=mouse", "--env", "b=cookie", "--env=empty", "--no-wait", "--revision-name=")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "created", "foo", "default"))

	output, err = executeServiceCommand(client, "update", "foo", "-e", "a=rabbit", "--env=empty-", "--no-wait", "--revision-name=")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "updated", "foo", "default"))

	r.Validate()
}

func TestServiceUpdateAnnotationsMock(t *testing.T) {
	client := clientservingv1.NewMockKnServiceClient(t)
	svcName := "svc1"
	newService := getService(svcName)
	template := &newService.Spec.Template
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	newService.ObjectMeta.Annotations = map[string]string{
		"an1": "staysConstant",
		"an2": "getsUpdated",
		"an3": "getsRemoved",
	}
	template.ObjectMeta.Annotations = map[string]string{
		"an1":                                "staysConstant",
		"an2":                                "getsUpdated",
		"an3":                                "getsRemoved",
		clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz",
	}

	updatedService := getService(svcName)
	template = &updatedService.Spec.Template
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	updatedService.ObjectMeta.Annotations = map[string]string{
		"an1": "staysConstant",
		"an2": "isUpdated",
	}
	template.ObjectMeta.Annotations = map[string]string{
		"an1":                                "staysConstant",
		"an2":                                "isUpdated",
		clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz",
	}

	r := client.Recorder()
	recordServiceUpdateWithSuccess(r, svcName, newService, updatedService)

	output, err := executeServiceCommand(client,
		"create", svcName, "--image", "gcr.io/foo/bar:baz",
		"--annotation", "an1=staysConstant",
		"--annotation", "an2=getsUpdated",
		"--annotation", "an3=getsRemoved",
		"--no-wait", "--revision-name=",
	)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "created", svcName, "default"))

	output, err = executeServiceCommand(client,
		"update", svcName,
		"--annotation", "an2=isUpdated",
		"--annotation", "an3-",
		"--no-wait", "--revision-name=",
	)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "updated", svcName, "default"))

	r.Validate()
}

func recordServiceUpdateWithSuccess(r *clientservingv1.ServingRecorder, svcName string, newService *servingv1.Service, updatedService *servingv1.Service) {
	r.GetService(svcName, nil, errors.NewNotFound(servingv1.Resource("service"), svcName))
	r.CreateService(newService, nil)
	r.GetService(svcName, newService, nil)
	r.UpdateService(updatedService, nil)
}

func TestServiceUpdateEnvFromAddingWithConfigMap(t *testing.T) {
	client := clientservingv1.NewMockKnServiceClient(t)
	svcName := "svc1"
	// prepare original state
	newService := getService(svcName)
	template := &newService.Spec.Template
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	template.Spec.Containers[0].EnvFrom = []corev1.EnvFromSource{
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "existing-name",
				},
			},
		},
	}

	// prepare updated state
	updatedService := getService(svcName)
	template = &updatedService.Spec.Template
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	template.Spec.Containers[0].EnvFrom = []corev1.EnvFromSource{
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "existing-name",
				},
			},
		},
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "new-name",
				},
			},
		},
	}

	r := client.Recorder()
	recordServiceUpdateWithSuccess(r, svcName, newService, updatedService)

	output, err := executeServiceCommand(client,
		"create", svcName, "--image", "gcr.io/foo/bar:baz",
		"--env-from", "config-map:existing-name",
		"--no-wait", "--revision-name=",
	)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "created", svcName, "default"))

	output, err = executeServiceCommand(client,
		"update", svcName,
		"--env-from", "config-map:new-name",
		"--no-wait", "--revision-name=",
	)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "updated", svcName, "default"))

	r.Validate()
}

func TestServiceUpdateEnvFromRemovalWithConfigMap(t *testing.T) {
	client := clientservingv1.NewMockKnServiceClient(t)
	svcName := "svc1"
	// prepare original state
	newService := getService(svcName)
	template := &newService.Spec.Template
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	template.Spec.Containers[0].EnvFrom = []corev1.EnvFromSource{
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "existing-name-1",
				},
			},
		},
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "existing-name-2",
				},
			},
		},
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "existing-name-3",
				},
			},
		},
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "existing-name-4",
				},
			},
		},
	}

	// prepare updated state
	updatedService1 := getService(svcName)
	template = &updatedService1.Spec.Template
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	template.Spec.Containers[0].EnvFrom = []corev1.EnvFromSource{
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "existing-name-4",
				},
			},
		},
	}

	// prepare updated state
	updatedService2 := getService(svcName)
	template = &updatedService2.Spec.Template
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	template.Spec.Containers[0].EnvFrom = []corev1.EnvFromSource{
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "existing-name-4",
				},
			},
		},
	}

	// prepare updated state
	updatedService3 := getService(svcName)
	template = &updatedService3.Spec.Template
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	template.Spec.Containers[0].EnvFrom = nil

	r := client.Recorder()
	recordServiceUpdateWithSuccess(r, svcName, newService, updatedService1)
	r.GetService(svcName, updatedService1, nil)
	//r.UpdateService(updatedService2, nil) // since an error happens, update is not triggered here
	r.GetService(svcName, updatedService2, nil)
	r.UpdateService(updatedService3, nil)

	output, err := executeServiceCommand(client,
		"create", svcName, "--image", "gcr.io/foo/bar:baz",
		"--env-from", "config-map:existing-name-1",
		"--env-from", "config-map:existing-name-2",
		"--env-from", "cm:existing-name-3",
		"--env-from", "config-map:existing-name-4",
		"--no-wait", "--revision-name=",
	)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "created", svcName, "default"))

	output, err = executeServiceCommand(client,
		"update", svcName,
		"--env-from", "config-map:existing-name-1-",
		"--env-from", "cm:existing-name-2-",
		"--env-from", "config-map:existing-name-3-",
		"--no-wait", "--revision-name=",
	)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "updated", svcName, "default"))

	// empty string
	output, err = executeServiceCommand(client,
		"update", svcName,
		"--env-from", "config-map:-",
		"--no-wait", "--revision-name=",
	)
	assert.Error(t, err, "the name of config-map cannot be an empty string")

	output, err = executeServiceCommand(client,
		"update", svcName,
		"--env-from", "cm:existing-name-4-",
		"--no-wait", "--revision-name=",
	)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "updated", svcName, "default"))

	r.Validate()
}

func TestServiceUpdateEnvFromRemovalWithEmptyName(t *testing.T) {
	client := clientservingv1.NewMockKnServiceClient(t)
	svcName := "svc1"
	// prepare original state
	newService := getService(svcName)
	template := &newService.Spec.Template
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	template.Spec.Containers[0].EnvFrom = []corev1.EnvFromSource{
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "existing-name-1",
				},
			},
		},
	}

	// prepare updated state
	updatedService1 := getService(svcName)
	template = &updatedService1.Spec.Template
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	template.Spec.Containers[0].EnvFrom = []corev1.EnvFromSource{
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "existing-name-1",
				},
			},
		},
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "existing-name-2",
				},
			},
		},
	}

	r := client.Recorder()
	r.GetService(svcName, nil, errors.NewNotFound(servingv1.Resource("service"), svcName))
	r.CreateService(newService, nil)
	r.GetService(svcName, newService, nil)
	r.GetService(svcName, newService, nil)
	r.UpdateService(updatedService1, nil)

	output, err := executeServiceCommand(client,
		"create", svcName, "--image", "gcr.io/foo/bar:baz",
		"--env-from", "config-map:existing-name-1",
		"--no-wait", "--revision-name=",
	)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "created", svcName, "default"))

	_, err = executeServiceCommand(client,
		"update", svcName,
		"--env-from", "-",
		"--no-wait", "--revision-name=",
	)
	assert.Error(t, err, "\"-\" is not a valid value for \"--env-from\"")

	output, err = executeServiceCommand(client,
		"update", svcName,
		"--env-from", "config-map:existing-name-2",
		"--no-wait", "--revision-name=",
	)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "updated", svcName, "default"))
}

func TestServiceUpdateEnvFromExistingWithConfigMap(t *testing.T) {
	client := clientservingv1.NewMockKnServiceClient(t)
	svcName := "svc1"
	// prepare original state
	newService := getService(svcName)
	template := &newService.Spec.Template
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	template.Spec.Containers[0].EnvFrom = []corev1.EnvFromSource{
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "existing-name-1",
				},
			},
		},
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "existing-name-2",
				},
			},
		},
	}

	// prepare updated state
	updatedService := getService(svcName)
	template = &updatedService.Spec.Template
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	template.Spec.Containers[0].EnvFrom = []corev1.EnvFromSource{
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "existing-name-1",
				},
			},
		},
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "existing-name-2",
				},
			},
		},
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "new-name",
				},
			},
		},
	}

	r := client.Recorder()
	recordServiceUpdateWithSuccess(r, svcName, newService, updatedService)

	output, err := executeServiceCommand(client,
		"create", svcName, "--image", "gcr.io/foo/bar:baz",
		"--env-from", "config-map:existing-name-1",
		"--env-from", "config-map:existing-name-2",
		"--no-wait", "--revision-name=",
	)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "created", svcName, "default"))

	output, err = executeServiceCommand(client,
		"update", svcName,
		"--env-from", "config-map:existing-name-1",
		"--env-from", "config-map:new-name",
		"--no-wait", "--revision-name=",
	)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "updated", svcName, "default"))

	r.Validate()
}

func TestServiceUpdateEnvFromAddingWithSecret(t *testing.T) {
	client := clientservingv1.NewMockKnServiceClient(t)
	svcName := "svc1"
	// prepare original state
	newService := getService(svcName)
	template := &newService.Spec.Template
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	template.Spec.Containers[0].EnvFrom = []corev1.EnvFromSource{
		{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "existing-name",
				},
			},
		},
	}

	// prepare updated state
	updatedService := getService(svcName)
	template = &updatedService.Spec.Template
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	template.Spec.Containers[0].EnvFrom = []corev1.EnvFromSource{
		{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "existing-name",
				},
			},
		},
		{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "new-name",
				},
			},
		},
	}

	r := client.Recorder()
	recordServiceUpdateWithSuccess(r, svcName, newService, updatedService)

	output, err := executeServiceCommand(client,
		"create", svcName, "--image", "gcr.io/foo/bar:baz",
		"--env-from", "secret:existing-name",
		"--no-wait", "--revision-name=",
	)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "created", svcName, "default"))

	output, err = executeServiceCommand(client,
		"update", svcName,
		"--env-from", "sc:new-name",
		"--no-wait", "--revision-name=",
	)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "updated", svcName, "default"))

	r.Validate()
}
func TestServiceUpdateEnvFromRemovalWithSecret(t *testing.T) {
	client := clientservingv1.NewMockKnServiceClient(t)
	svcName := "svc1"
	// prepare original state
	newService := getService(svcName)
	template := &newService.Spec.Template
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	template.Spec.Containers[0].EnvFrom = []corev1.EnvFromSource{
		{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "existing-name-1",
				},
			},
		},
		{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "existing-name-2",
				},
			},
		},
		{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "existing-name-3",
				},
			},
		},
		{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "existing-name-4",
				},
			},
		},
	}

	// prepare updated state
	updatedService1 := getService(svcName)
	template = &updatedService1.Spec.Template
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	template.Spec.Containers[0].EnvFrom = []corev1.EnvFromSource{
		{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "existing-name-4",
				},
			},
		},
	}

	// prepare updated state
	updatedService2 := getService(svcName)
	template = &updatedService2.Spec.Template
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	template.Spec.Containers[0].EnvFrom = []corev1.EnvFromSource{
		{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "existing-name-4",
				},
			},
		},
	}

	// prepare updated state
	updatedService3 := getService(svcName)
	template = &updatedService3.Spec.Template
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	template.Spec.Containers[0].EnvFrom = nil

	r := client.Recorder()
	r.GetService(svcName, nil, errors.NewNotFound(servingv1.Resource("service"), svcName))
	r.CreateService(newService, nil)
	r.GetService(svcName, newService, nil)
	r.UpdateService(updatedService1, nil)
	r.GetService(svcName, updatedService1, nil)
	//r.UpdateService(updatedService2, nil) // since an error happens, update is not triggered here
	r.GetService(svcName, updatedService2, nil)
	r.UpdateService(updatedService3, nil)

	output, err := executeServiceCommand(client,
		"create", svcName, "--image", "gcr.io/foo/bar:baz",
		"--env-from", "sc:existing-name-1",
		"--env-from", "secret:existing-name-2",
		"--env-from", "sc:existing-name-3",
		"--env-from", "secret:existing-name-4",
		"--no-wait", "--revision-name=",
	)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "created", svcName, "default"))

	output, err = executeServiceCommand(client,
		"update", svcName,
		"--env-from", "secret:existing-name-1-",
		"--env-from", "sc:existing-name-2-",
		"--env-from", "secret:existing-name-3-",
		"--no-wait", "--revision-name=",
	)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "updated", svcName, "default"))

	// empty string
	output, err = executeServiceCommand(client,
		"update", svcName,
		"--env-from", "secret:-",
		"--no-wait", "--revision-name=",
	)
	assert.Error(t, err, "the name of secret cannot be an empty string")

	output, err = executeServiceCommand(client,
		"update", svcName,
		"--env-from", "sc:existing-name-4-",
		"--no-wait", "--revision-name=",
	)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "updated", svcName, "default"))

	r.Validate()
}

func TestServiceUpdateEnvFromExistingWithSecret(t *testing.T) {
	client := clientservingv1.NewMockKnServiceClient(t)
	svcName := "svc1"
	// prepare original state
	newService := getService(svcName)
	template := &newService.Spec.Template
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	template.Spec.Containers[0].EnvFrom = []corev1.EnvFromSource{
		{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "existing-name-1",
				},
			},
		},
		{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "existing-name-2",
				},
			},
		},
	}

	// prepare updated state
	updatedService := getService(svcName)
	template = &updatedService.Spec.Template
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	template.Spec.Containers[0].EnvFrom = []corev1.EnvFromSource{
		{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "existing-name-1",
				},
			},
		},
		{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "existing-name-2",
				},
			},
		},
		{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "new-name",
				},
			},
		},
	}

	r := client.Recorder()
	recordServiceUpdateWithSuccess(r, svcName, newService, updatedService)

	output, err := executeServiceCommand(client,
		"create", svcName, "--image", "gcr.io/foo/bar:baz",
		"--env-from", "sc:existing-name-1",
		"--env-from", "secret:existing-name-2",
		"--no-wait", "--revision-name=",
	)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "created", svcName, "default"))

	output, err = executeServiceCommand(client,
		"update", svcName,
		"--env-from", "secret:existing-name-1",
		"--env-from", "secret:new-name",
		"--no-wait", "--revision-name=",
	)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "updated", svcName, "default"))

	r.Validate()
}

func TestServiceUpdateWithAddingVolume(t *testing.T) {
	client := clientservingv1.NewMockKnServiceClient(t)
	svcName := "svc1"
	// prepare original state
	newService := getService(svcName)
	template := &newService.Spec.Template
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	template.Spec.Volumes = []corev1.Volume{
		{
			Name: "vol-1",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "existing-config-map-1",
					},
				},
			},
		},
		{
			Name: "vol-2",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: "existing-secret-1",
				},
			},
		},
	}

	// prepare updated state
	updatedService := getService(svcName)
	template = &updatedService.Spec.Template
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	template.Spec.Volumes = []corev1.Volume{
		{
			Name: "vol-1",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "existing-config-map-1",
					},
				},
			},
		},
		{
			Name: "vol-2",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: "existing-secret-1",
				},
			},
		},
		{
			Name: "vol-3",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "existing-config-map-2",
					},
				},
			},
		},
		{
			Name: "vol-4",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: "existing-secret-2",
				},
			},
		},
	}

	r := client.Recorder()
	recordServiceUpdateWithSuccess(r, svcName, newService, updatedService)

	output, err := executeServiceCommand(client,
		"create", svcName, "--image", "gcr.io/foo/bar:baz",
		"--volume", "vol-1=cm:existing-config-map-1",
		"--volume", "vol-2=secret:existing-secret-1",
		"--no-wait", "--revision-name=",
	)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "created", svcName, "default"))

	output, err = executeServiceCommand(client,
		"update", svcName,
		"--volume", "vol-3=cm:existing-config-map-2",
		"--volume", "vol-4=secret:existing-secret-2",
		"--no-wait", "--revision-name=",
	)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "updated", svcName, "default"))

	r.Validate()
}

func TestServiceUpdateWithUpdatingVolume(t *testing.T) {
	client := clientservingv1.NewMockKnServiceClient(t)
	svcName := "svc1"
	// prepare original state
	newService := getService(svcName)
	template := &newService.Spec.Template
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	template.Spec.Volumes = []corev1.Volume{
		{
			Name: "vol-1",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "existing-config-map-1",
					},
				},
			},
		},
		{
			Name: "vol-2",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: "existing-secret-1",
				},
			},
		},
	}

	// prepare updated state
	updatedService := getService(svcName)
	template = &updatedService.Spec.Template
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	template.Spec.Volumes = []corev1.Volume{
		{
			Name: "vol-1",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "existing-config-map-3",
					},
				},
			},
		},
		{
			Name: "vol-2",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: "existing-secret-3",
				},
			},
		},
		{
			Name: "vol-3",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "existing-config-map-2",
					},
				},
			},
		},
		{
			Name: "vol-4",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: "existing-secret-2",
				},
			},
		},
	}

	r := client.Recorder()
	recordServiceUpdateWithSuccess(r, svcName, newService, updatedService)

	output, err := executeServiceCommand(client,
		"create", svcName, "--image", "gcr.io/foo/bar:baz",
		"--volume", "vol-1=cm:existing-config-map-1",
		"--volume", "vol-2=secret:existing-secret-1",
		"--no-wait", "--revision-name=",
	)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "created", svcName, "default"))

	output, err = executeServiceCommand(client,
		"update", svcName,
		"--volume", "vol-1=cm:existing-config-map-3",
		"--volume", "vol-2=secret:existing-secret-3",
		"--volume", "vol-3=cm:existing-config-map-2",
		"--volume", "vol-4=secret:existing-secret-2",
		"--no-wait", "--revision-name=",
	)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "updated", svcName, "default"))

	r.Validate()
}

func TestServiceUpdateWithRemovingVolume(t *testing.T) {
	client := clientservingv1.NewMockKnServiceClient(t)
	svcName := "svc1"
	// prepare original state
	newService := getService(svcName)
	template := &newService.Spec.Template
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	template.Spec.Volumes = []corev1.Volume{
		{
			Name: "vol-1",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "existing-config-map-1",
					},
				},
			},
		},
		{
			Name: "vol-2",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: "existing-secret-1",
				},
			},
		},
		{
			Name: "vol-3",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "existing-config-map-2",
					},
				},
			},
		},
		{
			Name: "vol-4",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: "existing-secret-2",
				},
			},
		},
	}

	// prepare updated state
	updatedService := getService(svcName)
	template = &updatedService.Spec.Template
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	template.Spec.Volumes = []corev1.Volume{
		{
			Name: "vol-1",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "existing-config-map-1",
					},
				},
			},
		},
		{
			Name: "vol-4",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: "existing-secret-2",
				},
			},
		},
	}

	r := client.Recorder()
	recordServiceUpdateWithSuccess(r, svcName, newService, updatedService)

	output, err := executeServiceCommand(client,
		"create", svcName, "--image", "gcr.io/foo/bar:baz",
		"--volume", "vol-1=cm:existing-config-map-1",
		"--volume", "vol-2=secret:existing-secret-1",
		"--volume", "vol-3=cm:existing-config-map-2",
		"--volume", "vol-4=secret:existing-secret-2",
		"--no-wait", "--revision-name=",
	)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "created", svcName, "default"))

	output, err = executeServiceCommand(client,
		"update", svcName,
		"--volume", "vol-3-",
		"--volume", "vol-2-",
		"--no-wait", "--revision-name=",
	)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "updated", svcName, "default"))

	r.Validate()
}

func TestServiceUpdateWithAddingMount(t *testing.T) {
	client := clientservingv1.NewMockKnServiceClient(t)
	svcName := "svc1"
	// prepare original state
	newService := getService(svcName)
	template := &newService.Spec.Template
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	// prepare updated state
	updatedService := getService(svcName)
	template = &updatedService.Spec.Template
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	template.Spec.Volumes = []corev1.Volume{
		{
			Name: clientserving.GenerateVolumeName("/mount/config-map-path"),
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "config-map-name",
					},
				},
			},
		},
		{
			Name: clientserving.GenerateVolumeName("/mount/secret-path"),
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: "secret-name",
				},
			},
		},
	}

	template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
		{
			Name:      clientserving.GenerateVolumeName("/mount/config-map-path"),
			MountPath: "/mount/config-map-path",
			ReadOnly:  true,
		},
		{
			Name:      clientserving.GenerateVolumeName("/mount/secret-path"),
			MountPath: "/mount/secret-path",
			ReadOnly:  true,
		},
	}

	r := client.Recorder()
	recordServiceUpdateWithSuccess(r, svcName, newService, updatedService)

	output, err := executeServiceCommand(client,
		"create", svcName, "--image", "gcr.io/foo/bar:baz",
		"--no-wait", "--revision-name=",
	)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "created", svcName, "default"))

	output, err = executeServiceCommand(client,
		"update", svcName,
		"--mount", "/mount/config-map-path=cm:config-map-name",
		"--mount", "/mount/secret-path=secret:secret-name",
		"--no-wait", "--revision-name=",
	)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "updated", svcName, "default"))

	r.Validate()
}

func TestServiceUpdateWithUpdatingMount(t *testing.T) {
	client := clientservingv1.NewMockKnServiceClient(t)
	svcName := "svc1"
	// prepare original state
	newService := getService(svcName)
	template := &newService.Spec.Template
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	template.Spec.Volumes = []corev1.Volume{
		{
			Name: clientserving.GenerateVolumeName("/mount/config-map-path"),
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "config-map-name-1",
					},
				},
			},
		},
		{
			Name: clientserving.GenerateVolumeName("/mount/secret-path"),
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: "secret-name-1",
				},
			},
		},
	}

	template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
		{
			Name:      clientserving.GenerateVolumeName("/mount/config-map-path"),
			MountPath: "/mount/config-map-path",
			ReadOnly:  true,
		},
		{
			Name:      clientserving.GenerateVolumeName("/mount/secret-path"),
			MountPath: "/mount/secret-path",
			ReadOnly:  true,
		},
	}

	// prepare updated state
	updatedService := getService(svcName)
	template = &updatedService.Spec.Template
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	template.Spec.Volumes = []corev1.Volume{
		{
			Name: clientserving.GenerateVolumeName("/mount/config-map-path"),
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "config-map-name-2",
					},
				},
			},
		},
		{
			Name: clientserving.GenerateVolumeName("/mount/secret-path"),
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: "secret-name-2",
				},
			},
		},
	}

	template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
		{
			Name:      clientserving.GenerateVolumeName("/mount/config-map-path"),
			MountPath: "/mount/config-map-path",
			ReadOnly:  true,
		},
		{
			Name:      clientserving.GenerateVolumeName("/mount/secret-path"),
			MountPath: "/mount/secret-path",
			ReadOnly:  true,
		},
	}

	r := client.Recorder()
	recordServiceUpdateWithSuccess(r, svcName, newService, updatedService)

	output, err := executeServiceCommand(client,
		"create", svcName, "--image", "gcr.io/foo/bar:baz",
		"--mount", "/mount/config-map-path=cm:config-map-name-1",
		"--mount", "/mount/secret-path=secret:secret-name-1",
		"--no-wait", "--revision-name=",
	)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "created", svcName, "default"))

	// Original orderedness should be kept after updating even though the orders
	// between updating flags are not the same with the original one
	output, err = executeServiceCommand(client,
		"update", svcName,
		"--mount", "/mount/secret-path=secret:secret-name-2",
		"--mount", "/mount/config-map-path=cm:config-map-name-2",
		"--no-wait", "--revision-name=",
	)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "updated", svcName, "default"))

	r.Validate()
}

func TestServiceUpdateWithRemovingMount(t *testing.T) {
	client := clientservingv1.NewMockKnServiceClient(t)
	svcName := "svc1"
	// prepare original state
	newService := getService(svcName)
	template := &newService.Spec.Template
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	template.Spec.Volumes = []corev1.Volume{
		{
			Name: clientserving.GenerateVolumeName("/mount/config-map-path-1"),
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "config-map-name-1",
					},
				},
			},
		},
		{
			Name: clientserving.GenerateVolumeName("/mount/secret-path-1"),
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: "secret-name-1",
				},
			},
		},
		{
			Name: clientserving.GenerateVolumeName("/mount/config-map-path-2"),
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "config-map-name-2",
					},
				},
			},
		},
		{
			Name: clientserving.GenerateVolumeName("/mount/secret-path-2"),
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: "secret-name-2",
				},
			},
		},
		{
			Name: "custom-vol",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "config-map",
					},
				},
			},
		},
	}

	template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
		{
			Name:      clientserving.GenerateVolumeName("/mount/config-map-path-1"),
			MountPath: "/mount/config-map-path-1",
			ReadOnly:  true,
		},
		{
			Name:      clientserving.GenerateVolumeName("/mount/secret-path-1"),
			MountPath: "/mount/secret-path-1",
			ReadOnly:  true,
		},
		{
			Name:      clientserving.GenerateVolumeName("/mount/config-map-path-2"),
			MountPath: "/mount/config-map-path-2",
			ReadOnly:  true,
		},
		{
			Name:      clientserving.GenerateVolumeName("/mount/secret-path-2"),
			MountPath: "/mount/secret-path-2",
			ReadOnly:  true,
		},
		{
			Name:      "custom-vol",
			MountPath: "/mount/custom-path",
			ReadOnly:  true,
		},
	}

	// prepare updated state
	updatedService1 := getService(svcName)
	template = &updatedService1.Spec.Template
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	template.Spec.Volumes = []corev1.Volume{
		{
			Name: clientserving.GenerateVolumeName("/mount/config-map-path-1"),
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "config-map-name-1",
					},
				},
			},
		},
		{
			Name: clientserving.GenerateVolumeName("/mount/secret-path-2"),
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: "secret-name-2",
				},
			},
		},
		{
			Name: "custom-vol",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "config-map",
					},
				},
			},
		},
	}

	template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
		{
			Name:      clientserving.GenerateVolumeName("/mount/config-map-path-1"),
			MountPath: "/mount/config-map-path-1",
			ReadOnly:  true,
		},
		{
			Name:      clientserving.GenerateVolumeName("/mount/secret-path-2"),
			MountPath: "/mount/secret-path-2",
			ReadOnly:  true,
		},
		{
			Name:      "custom-vol",
			MountPath: "/mount/custom-path",
			ReadOnly:  true,
		},
	}

	// prepare updated state
	updatedService2 := getService(svcName)
	template = &updatedService2.Spec.Template
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	r := client.Recorder()
	r.GetService(svcName, nil, errors.NewNotFound(servingv1.Resource("service"), svcName))
	r.CreateService(newService, nil)
	r.GetService(svcName, newService, nil)
	r.UpdateService(updatedService1, nil)
	r.GetService(svcName, updatedService1, nil)
	r.UpdateService(updatedService2, nil)

	output, err := executeServiceCommand(client,
		"create", svcName, "--image", "gcr.io/foo/bar:baz",
		"--mount", "/mount/config-map-path-1=cm:config-map-name-1",
		"--mount", "/mount/secret-path-1=secret:secret-name-1",
		"--mount", "/mount/config-map-path-2=cm:config-map-name-2",
		"--mount", "/mount/secret-path-2=secret:secret-name-2",
		"--mount", "/mount/custom-path=custom-vol",
		"--volume", "custom-vol=cm:config-map",
		"--no-wait", "--revision-name=",
	)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "created", svcName, "default"))

	template.Spec.Volumes = []corev1.Volume{
		{
			Name: "custom-vol",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "config-map",
					},
				},
			},
		},
	}

	output, err = executeServiceCommand(client,
		"update", svcName,
		"--mount", "/mount/config-map-path-2-",
		"--mount", "/mount/secret-path-1-",
		"--no-wait", "--revision-name=",
	)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "updated", svcName, "default"))

	output, err = executeServiceCommand(client,
		"update", svcName,
		"--mount", "/mount/config-map-path-1-",
		"--mount", "/mount/secret-path-2-",
		"--mount", "/mount/custom-path-",
		"--no-wait", "--revision-name=",
	)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "updated", svcName, "default"))

	r.Validate()
}

func TestServiceUpdateUser(t *testing.T) {
	client := clientservingv1.NewMockKnServiceClient(t)
	svcName := "svc1"
	newService := getService(svcName)
	template := &newService.Spec.Template
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{
		RunAsUser: ptr.Int64(int64(1001)),
	}
	template.ObjectMeta.Annotations = map[string]string{
		clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz",
	}

	updatedService := getService(svcName)
	template = &updatedService.Spec.Template
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{
		RunAsUser: ptr.Int64(int64(1002)),
	}
	template.ObjectMeta.Annotations = map[string]string{
		clientserving.UserImageAnnotationKey: "gcr.io/foo/bar:baz",
	}

	r := client.Recorder()
	recordServiceUpdateWithSuccess(r, svcName, newService, updatedService)

	output, err := executeServiceCommand(client,
		"create", svcName, "--image", "gcr.io/foo/bar:baz",
		"--user", "1001",
		"--no-wait", "--revision-name=",
	)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "created", svcName, "default"))

	output, err = executeServiceCommand(client,
		"update", svcName,
		"--user", "1002",
		"--no-wait", "--revision-name=",
	)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "updated", svcName, "default"))

	r.Validate()
}
