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
	"fmt"

	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

func ContainerOfRevisionTemplate(template *servingv1alpha1.RevisionTemplateSpec) (*corev1.Container, error) {
	return ContainerOfRevisionSpec(&template.Spec)
}

func ContainerOfRevisionSpec(revisionSpec *servingv1alpha1.RevisionSpec) (*corev1.Container, error) {
	if usesOldV1alpha1ContainerField(revisionSpec) {
		return revisionSpec.DeprecatedContainer, nil
	}
	containers := revisionSpec.Containers
	if len(containers) == 0 {
		return nil, fmt.Errorf("internal: no container set in spec.template.spec.containers")
	}
	if len(containers) > 1 {
		return nil, fmt.Errorf("internal: can't extract container for updating environment"+
			" variables as the configuration contains "+
			"more than one container (i.e. %d containers)", len(containers))
	}
	return &containers[0], nil
}

// =======================================================================================

func usesOldV1alpha1ContainerField(revisionSpec *servingv1alpha1.RevisionSpec) bool {
	return revisionSpec.DeprecatedContainer != nil
}
