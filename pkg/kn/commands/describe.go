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
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	"knative.dev/client/pkg/printers"
	"knative.dev/pkg/apis"
)

// Max length When to truncate long strings (when not "all" mode switched on)
const TruncateAt = 100

func WriteMetadata(dw printers.PrefixWriter, m *metav1.ObjectMeta, printDetails bool) {
	dw.WriteAttribute("Name", m.Name)
	dw.WriteAttribute("Namespace", m.Namespace)
	WriteMapDesc(dw, m.Labels, "Labels", printDetails)
	WriteMapDesc(dw, m.Annotations, "Annotations", printDetails)
	dw.WriteAttribute("Age", Age(m.CreationTimestamp.Time))
}

var boringDomains = map[string]bool{
	"serving.knative.dev":   true,
	"client.knative.dev":    true,
	"kubectl.kubernetes.io": true,
}

func keyIsBoring(k string) bool {
	parts := strings.Split(k, "/")
	return len(parts) > 1 && boringDomains[parts[0]]
}

// Write a map either compact in a single line (possibly truncated) or, if printDetails is set,
// over multiple line, one line per key-value pair. The output is sorted by keys.
func WriteMapDesc(dw printers.PrefixWriter, m map[string]string, label string, details bool) {
	if len(m) == 0 {
		return
	}

	var keys []string
	for k := range m {
		if details || !keyIsBoring(k) {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		return
	}
	sort.Strings(keys)

	if details {
		for i, key := range keys {
			l := ""
			if i == 0 {
				l = printers.Label(label)
			}
			dw.WriteColsLn(l, key+"="+m[key])
		}
		return
	}

	dw.WriteColsLn(printers.Label(label), joinAndTruncate(keys, m, TruncateAt-len(label)-2))
}

func Age(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return duration.ShortHumanDuration(time.Now().Sub(t))
}

func formatConditionType(condition apis.Condition) string {
	return string(condition.Type)
}

// Status in ASCII format
func formatStatus(c apis.Condition) string {
	switch c.Status {
	case corev1.ConditionTrue:
		return "++"
	case corev1.ConditionFalse:
		switch c.Severity {
		case apis.ConditionSeverityError:
			return "!!"
		case apis.ConditionSeverityWarning:
			return " W"
		case apis.ConditionSeverityInfo:
			return " I"
		default:
			return " !"
		}
	default:
		return "??"
	}
}

// Used for conditions table to do own formatting for the table,
// as the tabbed writer doesn't work nicely with colors
func getMaxTypeLen(conditions []apis.Condition) int {
	max := 0
	for _, condition := range conditions {
		if len(condition.Type) > max {
			max = len(condition.Type)
		}
	}
	return max
}

// Sort conditions: Ready first, followed by error, then Warning, then Info
func sortConditions(conditions []apis.Condition) []apis.Condition {
	// Don't change the orig slice
	ret := make([]apis.Condition, len(conditions))
	for i, c := range conditions {
		ret[i] = c
	}
	sort.SliceStable(ret, func(i, j int) bool {
		ic := &ret[i]
		jc := &ret[j]
		// Ready first
		if ic.Type == apis.ConditionReady {
			return jc.Type != apis.ConditionReady
		}
		if jc.Type == apis.ConditionReady {
			return false
		}
		// Among conditions of the same Severity, sort by Type
		if ic.Severity == jc.Severity {
			return ic.Type < jc.Type
		}
		// Error < Warning < Info
		switch ic.Severity {
		case apis.ConditionSeverityError:
			return true
		case apis.ConditionSeverityWarning:
			return jc.Severity == apis.ConditionSeverityInfo
		case apis.ConditionSeverityInfo:
			return false
		default:
			return false
		}
	})
	return ret
}

// Print out a table with conditions.
func WriteConditions(dw printers.PrefixWriter, conditions []apis.Condition, printMessage bool) {
	section := dw.WriteAttribute("Conditions", "")
	conditions = sortConditions(conditions)
	maxLen := getMaxTypeLen(conditions)
	formatHeader := "%-2s %-" + strconv.Itoa(maxLen) + "s %6s %-s\n"
	formatRow := "%-2s %-" + strconv.Itoa(maxLen) + "s %6s %-s\n"
	section.Writef(formatHeader, "OK", "TYPE", "AGE", "REASON")
	for _, condition := range conditions {
		ok := formatStatus(condition)
		reason := condition.Reason
		if printMessage && reason != "" {
			reason = fmt.Sprintf("%s (%s)", reason, condition.Message)
		}
		section.Writef(formatRow, ok, formatConditionType(condition), Age(condition.LastTransitionTime.Inner.Time), reason)
	}
}

// Writer a slice compact (printDetails == false) in one line, or over multiple line
// with key-value line-by-line (printDetails == true)
func WriteSliceDesc(dw printers.PrefixWriter, s []string, label string, printDetails bool) {

	if len(s) == 0 {
		return
	}

	if printDetails {
		for i, value := range s {
			if i == 0 {
				dw.WriteColsLn(printers.Label(label), value)
			} else {
				dw.WriteColsLn("", value)
			}
		}
		return
	}

	joined := strings.Join(s, ", ")
	if len(joined) > TruncateAt {
		joined = joined[:TruncateAt-4] + " ..."
	}
	dw.WriteAttribute(label, joined)
}

// Join to key=value pair, comma separated, and truncate if longer than a limit
func joinAndTruncate(sortedKeys []string, m map[string]string, width int) string {
	ret := ""
	for _, key := range sortedKeys {
		ret += fmt.Sprintf("%s=%s, ", key, m[key])
		if len(ret) > width {
			break
		}
	}
	// cut of two latest chars
	ret = strings.TrimRight(ret, ", ")
	if len(ret) <= width {
		return ret
	}
	return string(ret[:width-4]) + " ..."
}
