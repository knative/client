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
	"fmt"
	"github.com/knative/client/pkg/wait"
	"github.com/knative/pkg/apis"
	serving_v1alpha1_api "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	serving_v1alpha1_client "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Create wait arguments for a Knative service which can be used to wait for
// a create/update options to be finished
// Can be used by `service_create` and `service_update`, hence this extra file
func newServiceWaitForReady(client serving_v1alpha1_client.ServingV1alpha1Interface, namespace string) wait.WaitForReady {
	return wait.NewWaitForReady(
		"service",
		client.Services(namespace).Watch,
		serviceConditionExtractor)
}

func serviceConditionExtractor(obj runtime.Object) (apis.Conditions, error) {
	service, ok := obj.(*serving_v1alpha1_api.Service)
	if !ok {
		return nil, fmt.Errorf("%v is not a service", obj)
	}
	return apis.Conditions(service.Status.Conditions), nil
}
