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
