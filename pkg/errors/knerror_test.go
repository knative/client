package errors

import (
	"gotest.tools/assert"
	"testing"
)

func TestNewKNError(t *testing.T) {
	err := NewKNError("myerror")
	assert.Error(t, err, "myerror")

	err = NewKNError("")
	assert.Error(t, err, "")
}

func TestKNError_Error(t *testing.T) {
	err := NewKNError("myerror")
	assert.Equal(t, err.Error(), "myerror")

	err = NewKNError("")
	assert.Equal(t, err.Error(), "")
}
