package errors

import (
	"gotest.tools/assert"
	"testing"
)

func TestNewInvalidCRD(t *testing.T) {
	err := newInvalidCRD("serving.knative.dev")
	assert.Error(t, err, "no Knative serving API found on the backend. Please verify the installation.")

	err = newInvalidCRD("serving")
	assert.Error(t, err, "no Knative serving API found on the backend. Please verify the installation.")

	err = newInvalidCRD("")
	assert.Error(t, err, "no Knative  API found on the backend. Please verify the installation.")

}
