package v1alpha1

import (
	"testing"

	"knative.dev/pkg/apis/duck/v1beta1"
)

func TestMockKnClient(t *testing.T) {

	client := NewMockKnCronJobSourceClient(t)

	recorder := client.Recorder()

	// Record all services
	recorder.GetCronJobSource("hello", nil, nil)
	recorder.CreateCronJobSource("hello", "* * * * *","",&v1beta1.Destination{}, nil)
	recorder.UpdateCronJobSource("hello", "* * * * *","",&v1beta1.Destination{}, nil)
	recorder.DeleteCronJobSource("hello", nil)

	// Call all service
	client.GetCronJobSource("hello")
	client.CreateCronJobSource("hello", "* * * * *", "", &v1beta1.Destination{})
	client.UpdateCronJobSource("hello", "* * * * *", "", &v1beta1.Destination{})
	client.DeleteCronJobSource("hello")

	// Validate
	recorder.Validate()
}
