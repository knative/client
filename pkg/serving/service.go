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
	"bytes"
	"errors"
	"math/rand"
	"strings"
	"text/template"

	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
)

// Get the revision template associated with a service.
// Depending on the structure returned either the new v1beta1 fields or the
// 'old' v1alpha1 fields are looked up.
// The returned revision template can be updated in place.
// An error is returned if no revision template could be extracted
func RevisionTemplateOfService(service *servingv1alpha1.Service) (*servingv1alpha1.RevisionTemplateSpec, error) {
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

var charChoices = []string{
	"b", "c", "d", "f", "g", "h", "j", "k", "l", "m", "n", "p", "q", "r", "s", "t", "v", "w", "x",
	"y", "z",
}

type revisionTemplContext struct {
	Service    string
	Generation int64
}

func (c *revisionTemplContext) Random(l int) string {
	chars := make([]string, 0, l)
	for i := 0; i < l; i++ {
		chars = append(chars, charChoices[rand.Int()%len(charChoices)])
	}
	return strings.Join(chars, "")
}

// GenerateRevisionName returns an automatically-generated name suitable for the
// next revision of the given service.
func GenerateRevisionName(nameTempl string, service *servingv1alpha1.Service) (string, error) {
	templ, err := template.New("revisionName").Parse(nameTempl)
	if err != nil {
		return "", err
	}
	context := &revisionTemplContext{
		Service:    service.Name,
		Generation: service.Generation + 1,
	}
	buf := new(bytes.Buffer)
	err = templ.Execute(buf, context)
	if err != nil {
		return "", err
	}
	res := buf.String()
	// Empty is ok.
	if res == "" {
		return res, nil
	}
	prefix := service.Name + "-"
	if !strings.HasPrefix(res, prefix) {
		res = prefix + res
	}
	return res, nil
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
