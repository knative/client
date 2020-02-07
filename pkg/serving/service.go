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
	"math/rand"
	"strings"
	"text/template"

	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

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
func GenerateRevisionName(nameTempl string, service *servingv1.Service) (string, error) {
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
