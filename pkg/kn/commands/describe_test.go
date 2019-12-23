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

package commands

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"gotest.tools/assert"
	"knative.dev/client/pkg/printers"
	"knative.dev/pkg/apis"
)

var testMap = map[string]string{
	"a":                         "b",
	"c":                         "d",
	"foo":                       "bar",
	"serving.knative.dev/funky": "chicken",
}

func TestWriteMapDesc(t *testing.T) {
	buf := &bytes.Buffer{}
	dw := printers.NewBarePrefixWriter(buf)
	WriteMapDesc(dw, testMap, "eggs", false)
	assert.Equal(t, buf.String(), "eggs:\ta=b, c=d, foo=bar\n")
}

func TestWriteMapDescDetailed(t *testing.T) {
	buf := &bytes.Buffer{}
	dw := printers.NewBarePrefixWriter(buf)
	WriteMapDesc(dw, testMap, "eggs", true)
	assert.Equal(t, buf.String(), "eggs:\ta=b\n\tc=d\n\tfoo=bar\n\tserving.knative.dev/funky=chicken\n")
}

func TestWriteMapTruncated(t *testing.T) {
	buf := &bytes.Buffer{}
	dw := printers.NewBarePrefixWriter(buf)

	items := map[string]string{}
	for i := 0; i < 1000; i++ {
		items[strconv.Itoa(i)] = strconv.Itoa(i + 1)
	}
	WriteMapDesc(dw, items, "eggs", false)
	assert.Assert(t, len(strings.TrimSpace(buf.String())) <= TruncateAt)
}

var someConditions = []apis.Condition{
	{Type: apis.ConditionReady, Status: "True"},
	{Type: "Aaa", Status: "True"},
	{Type: "Zzz", Status: "False"},
	{Type: "Bbb", Status: "False", Severity: apis.ConditionSeverityWarning, Reason: "Bad"},
	{Type: "Ccc", Status: "False", Severity: apis.ConditionSeverityInfo, Reason: "Eh."},
}
var permutations = [][]int{
	{0, 1, 2, 3, 4},
	{4, 3, 2, 1, 0},
	{2, 1, 4, 3, 0},
	{2, 1, 0, 3, 4},
}

func TestSortConditions(t *testing.T) {
	for _, p := range permutations {
		permuted := make([]apis.Condition, len(someConditions))
		for i, j := range p {
			permuted[i] = someConditions[j]
		}
		sorted := sortConditions(permuted)
		assert.DeepEqual(t, sorted, someConditions)
	}
}

var spaces = regexp.MustCompile("\\s*")

func normalizeSpace(s string) string {
	return spaces.ReplaceAllLiteralString(s, " ")
}

func TestWriteConditions(t *testing.T) {
	for _, p := range permutations {
		permuted := make([]apis.Condition, len(someConditions))
		for i, j := range p {
			permuted[i] = someConditions[j]
		}
		buf := &bytes.Buffer{}
		dw := printers.NewBarePrefixWriter(buf)
		WriteConditions(dw, permuted, false)
		assert.Equal(t, normalizeSpace(buf.String()), normalizeSpace(`Conditions:
OK TYPE AGE REASON
++ Ready
++ Aaa
!! Zzz
 W Bbb Bad
 I Ccc Eh.`))
	}
}
