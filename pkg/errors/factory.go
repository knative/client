package errors

import (
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func isCRDError(status api_errors.APIStatus) bool {
	for _, cause := range status.Status().Details.Causes {
		if strings.HasPrefix(cause.Message, "404")  && cause.Type == v1.CauseTypeUnexpectedServerResponse {
			return true
		}
	}

	return false
}

func Build(err error) error {
	apiStatus, ok := err.(api_errors.APIStatus)
	if !ok {
		return err
	}

	var knerr *KNError

	if isCRDError(apiStatus) {
		knerr = newInvalidCRD(apiStatus.Status().Details.Group)
		knerr.Status = apiStatus
		return knerr
	}

	return err
}
