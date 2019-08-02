package errors

import (
	"fmt"
	"strings"
)

func newInvalidCRD(apiGroup string) *KNError {
	parts := strings.Split(apiGroup, ".")
	name := parts[0]
	msg := fmt.Sprintf("no Knative %s API found on the backend. Please verify the installation.", name)

	return NewKNError(msg)
}
