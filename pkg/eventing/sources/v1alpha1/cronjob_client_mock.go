package v1alpha1

import (
	"testing"

	"knative.dev/eventing/pkg/apis/sources/v1alpha1"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"

	"knative.dev/client/pkg/util/mock"
)

type MockKnCronJobSourceClient struct {
	t         *testing.T
	recorder  *CronJobSourcesRecorder
	namespace string
}

// NewMockKnServiceClient returns a new mock instance which you need to record for
func NewMockKnCronJobSourceClient(t *testing.T, ns ...string) *MockKnCronJobSourceClient {
	namespace := "default"
	if len(ns) > 0 {
		namespace = ns[0]
	}
	return &MockKnCronJobSourceClient{
		t:        t,
		recorder: &CronJobSourcesRecorder{mock.NewRecorder(t, namespace)},
	}
}

// recorder for service
type CronJobSourcesRecorder struct {
	r *mock.Recorder
}

// Get the record to start for the recorder
func (c *MockKnCronJobSourceClient) Recorder() *CronJobSourcesRecorder {
	return c.recorder
}

// Namespace of this client
func (c *MockKnCronJobSourceClient) Namespace() string {
	return c.recorder.r.Namespace()
}

// Create CronJob
func (sr *CronJobSourcesRecorder) CreateCronJobSource(name, schedule, data, sink interface{}, err error) {
	sr.r.Add("CreateCronJobSource", []interface{}{name, schedule, data, sink}, []interface{}{err})
}

func (c *MockKnCronJobSourceClient) CreateCronJobSource(name, schedule, data string, sink *duckv1beta1.Destination) error {
	call := c.recorder.r.VerifyCall("CreateCronJobSource", name, schedule, data, sink)
	return mock.ErrorOrNil(call.Result[0])
}

// Get CronJob
func (sr *CronJobSourcesRecorder) GetCronJobSource(name interface{}, cronjobSource *v1alpha1.CronJobSource, err error) {
	sr.r.Add("GetCronJobSource", []interface{}{name}, []interface{}{cronjobSource, err})
}

func (c *MockKnCronJobSourceClient) GetCronJobSource(name string) (*v1alpha1.CronJobSource, error) {
	call := c.recorder.r.VerifyCall("GetCronJobSource", name)
	return call.Result[0].(*v1alpha1.CronJobSource), mock.ErrorOrNil(call.Result[1])
}

// Update CronJob
func (sr *CronJobSourcesRecorder) UpdateCronJobSource(name, schedule, data, sink interface{}, err error) {
	sr.r.Add("UpdateCronJobSource", []interface{}{name, schedule, data, sink}, []interface{}{err})
}

func (c *MockKnCronJobSourceClient) UpdateCronJobSource(name, schedule, data string, sink *duckv1beta1.Destination) error {
	call := c.recorder.r.VerifyCall("UpdateCronJobSource", name, schedule, data, sink)
	return mock.ErrorOrNil(call.Result[0])
}

// Delete CronJob
func (sr *CronJobSourcesRecorder) DeleteCronJobSource(name interface{}, err error) {
	sr.r.Add("DeleteCronJobSource", []interface{}{name}, []interface{}{err})
}

func (c *MockKnCronJobSourceClient) DeleteCronJobSource(name string) error {
	call := c.recorder.r.VerifyCall("DeleteCronJobSource", name)
	return mock.ErrorOrNil(call.Result[0])
}

// Check that every recorded method has been called
func (sr *CronJobSourcesRecorder) Validate() {
	sr.r.CheckThatAllRecordedMethodsHaveBeenCalled()
}
