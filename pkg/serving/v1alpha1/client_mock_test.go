package v1alpha1

import (
	"testing"
	"time"

	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
)

func TestMockKnClient(t *testing.T) {

	client := NewMockKnClient(t)

	recorder := client.Recorder()

	// Record all services
	recorder.GetService("hello", nil, nil)
	recorder.ListServices(Any(), nil, nil)
	recorder.CreateService(&v1alpha1.Service{},nil)
	recorder.UpdateService(&v1alpha1.Service{}, nil)
	recorder.DeleteService("hello", nil)
	recorder.WaitForService("hello", time.Duration(10) * time.Second, nil)
	recorder.GetRevision("hello", nil, nil)
	recorder.ListRevisions(Any(), nil, nil)
	recorder.DeleteRevision("hello", nil)
	recorder.GetRoute("hello", nil, nil)
	recorder.ListRoutes(Any(), nil, nil)

	// Call all services
	client.GetService("hello")
	client.ListServices(WithName("blub"))
	client.CreateService(&v1alpha1.Service{})
	client.UpdateService(&v1alpha1.Service{})
	client.DeleteService("hello")
	client.WaitForService("hello", time.Duration(10) * time.Second)
	client.GetRevision("hello")
	client.ListRevisions(WithName("blub"))
	client.DeleteRevision("hello")
	client.GetRoute("hello")
	client.ListRoutes(WithName("blub"))


	// Validate
	recorder.Validate()
}