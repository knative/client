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
	WriteMapDesc(dw, m.Labels, l("Labels"), "", printDetails)
	WriteMapDesc(dw, m.Annotations, l("Annotations"), "", printDetails)
	dw.WriteAttribute("Age", Age(m.CreationTimestamp.Time))

}

var boringDomains = map[string]bool{
	"serving.knative.dev":   true,
	"client.knative.dev":    true,
	"kubectl.kubernetes.io": true,
}

// Write a map either compact in a single line (possibly truncated) or, if printDetails is set,
// over multiple line, one line per key-value pair. The output is sorted by keys.
func WriteMapDesc(dw printers.PrefixWriter, m map[string]string, label string, labelPrefix string, details bool) {
	if len(m) == 0 {
		return
	}

	var keys []string
	for k := range m {
		parts := strings.Split(k, "/")
		if details || len(parts) <= 1 || !boringDomains[parts[0]] {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		return
	}
	sort.Strings(keys)

	if details {
		l := labelPrefix + label

		for _, key := range keys {
			dw.WriteColsLn(l, key+"="+m[key])
			l = labelPrefix
		}
		return
	}

	dw.WriteColsLn(label, joinAndTruncate(keys, m))
}

func Age(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return duration.ShortHumanDuration(time.Now().Sub(t))
}

// Color the type of the conditions
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
	ret := make([]apis.Condition, 0, len(conditions))
	for _, c := range conditions {
		ret = append(ret, c)
	}
	sort.SliceStable(ret, func(i, j int) bool {
		ic := &ret[i]
		jc := &ret[j]
		// Ready first
		if ic.Type == apis.ConditionReady {
			return jc.Type != apis.ConditionReady
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
		return false
	})
	return ret
}

// Print out a table with conditions. Use green for 'ok', and red for 'nok' if color is enabled
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

// Format label (extracted so that color could be added more easily to all labels)
func l(label string) string {
	return label + ":"
}

// Join to key=value pair, comma separated, and truncate if longer than a limit
func joinAndTruncate(sortedKeys []string, m map[string]string) string {
	ret := ""
	for _, key := range sortedKeys {
		ret += fmt.Sprintf("%s=%s, ", key, m[key])
		if len(ret) > TruncateAt {
			break
		}
	}
	// cut of two latest chars
	ret = strings.TrimRight(ret, ", ")
	if len(ret) <= TruncateAt {
		return ret
	}
	return string(ret[:TruncateAt-4]) + " ..."
}
