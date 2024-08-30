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
	"context"
	"errors"
	"flag"
	"net/http"
	"path/filepath"
	"strings"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
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

func isStatusError(err error) bool {
	var errAPIStatus api_errors.APIStatus
	return errors.As(err, &errAPIStatus)
}

func newStatusError(err error) error {
	var errAPIStatus api_errors.APIStatus
	errors.As(err, &errAPIStatus)

	if errAPIStatus.Status().Details == nil {
		return err
	}
	canFindResource := "unknown"
	var knerr *KNError
	if isCRDError(errAPIStatus) {
		if strings.Contains(errAPIStatus.Status().Message, "the server could not find") {
			resourceName := getResourceNameFromErrMessage(errAPIStatus.Status().Message)
			var kubeconfig *string
			if home := homedir.HomeDir(); home != "" {
				kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
			} else {
				kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
			}
			flag.Parse()
			// use the current context in kubeconfig
			config, _ := clientcmd.BuildConfigFromFlags("", *kubeconfig)

			if config != nil {
				canFindResource = "false"
				// instantiate our client with config
				clientset, err := apiextensionsv1.NewForConfig(config)
				if err != nil {
					return err
				}
				options := metav1.ListOptions{}
				crdList, err := clientset.CustomResourceDefinitions().List(context.Background(), options)
				if err != nil {
					return err
				}
				for _, crd := range crdList.Items {
					if crd.Name == resourceName {
						canFindResource = "true"
						break
					}
				}
			}
		}
		knerr = NewInvalidCRD(errAPIStatus.Status().Details.Group, canFindResource)
		knerr.Status = errAPIStatus
		return knerr
	}
	return err
}

func getResourceNameFromErrMessage(msg string) string {
	res1 := strings.SplitAfter(msg, "(get ")
	return res1[1][:len(res1[1])-1]
}

// Retrieves a custom error struct based on the original error APIStatus struct
// Returns the original error struct in case it can't identify the kind of APIStatus error
func GetError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case isStatusError(err):
		return newStatusError(err)
	case isEmptyConfigError(err):
		return newNoKubeConfig(err.Error())
	case isNoRouteToHostError(err):
		return newNoRouteToHost(err.Error())
	default:
		return err
	}
}

// IsForbiddenError returns true if given error can be converted to API status and of type forbidden access else false
func IsForbiddenError(err error) bool {
	var errAPIStatus api_errors.APIStatus
	if !errors.As(err, &errAPIStatus) {
		return false
	}
	return errAPIStatus.Status().Code == int32(http.StatusForbidden)
}
