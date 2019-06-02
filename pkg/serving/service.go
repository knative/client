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

package serving

import (
	"errors"

	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
)

// Get the revision template associated with a service.
// Depending on the structure returned either the new v1beta1 fields or the
// 'old' v1alpha1 fields are looked up.
// The returned revision template can be updated in place.
// An error is returned if no revision template could be extracted
func GetRevisionTemplate(service *servingv1alpha1.Service) (*servingv1alpha1.RevisionTemplateSpec, error) {
	// Try v1beta1 field first
	if service.Spec.Template != nil {
		return service.Spec.Template, nil
	}

	config, err := getConfiguration(service)
	if err != nil {
		return nil, err
	}
	return config.DeprecatedRevisionTemplate, nil
}

func getConfiguration(service *servingv1alpha1.Service) (*servingv1alpha1.ConfigurationSpec, error) {
	if service.Spec.DeprecatedRunLatest != nil {
		return &service.Spec.DeprecatedRunLatest.Configuration, nil
	} else if service.Spec.DeprecatedRelease != nil {
		return &service.Spec.DeprecatedRelease.Configuration, nil
	} else if service.Spec.DeprecatedPinned != nil {
		return &service.Spec.DeprecatedPinned.Configuration, nil
	} else {
		return nil, errors.New("service does not specify a Configuration")
	}
}
