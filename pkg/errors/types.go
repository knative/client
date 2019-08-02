package errors

import api_errors "k8s.io/apimachinery/pkg/api/errors"

type KNError struct {
	Status api_errors.APIStatus
	msg string
}