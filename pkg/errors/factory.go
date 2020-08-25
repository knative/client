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

package errors

import (
	"net/http"
	"strings"

	api_errors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func isCRDError(status api_errors.APIStatus) bool {
	for _, cause := range status.Status().Details.Causes {
		if strings.HasPrefix(cause.Message, "404") && cause.Type == v1.CauseTypeUnexpectedServerResponse {
			return true
		}
	}
	return false
}

func isNoRouteToHostError(err error) bool {
	return strings.Contains(err.Error(), "no route to host") || strings.Contains(err.Error(), "i/o timeout")
}

func isEmptyConfigError(err error) bool {
	return strings.Contains(err.Error(), "no configuration has been provided")
}

//Retrieves a custom error struct based on the original error APIStatus struct
//Returns the original error struct in case it can't identify the kind of APIStatus error
func GetError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case isEmptyConfigError(err):
		return newNoKubeConfig(err.Error())
	case isNoRouteToHostError(err):
		return newNoRouteToHost(err.Error())
	default:
		apiStatus, ok := err.(api_errors.APIStatus)
		if !ok {
			return err
		}
		if apiStatus.Status().Details == nil {
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
}

// IsForbiddenError returns true if given error can be converted to API status and of type forbidden access else false
func IsForbiddenError(err error) bool {
	if status, ok := err.(api_errors.APIStatus); ok {
		return status.Status().Code == int32(http.StatusForbidden)
	}
	return false
}
