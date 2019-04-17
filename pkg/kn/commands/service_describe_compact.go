package commands

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/knative/client/pkg/printers"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"io"
	v1 "k8s.io/api/core/v1"
	"strconv"
	"strings"
)

var emojiMap = map[string]string{
	"Name":        "🆔",
	"Namespace":   "📁",
	"URL":         "🌍",
	"Address":     "🏠",
	"Annotations": "✏️",
	"Labels":      "🏷️",
	"Age":         "⏰",
	"Revisions":   "❄️",
	"Image":       "📦",
	"Env":         "🌵",
	"Port":        "🔌",
	"Generation":  "📶",
}

func describeCompact(w io.Writer, service *v1alpha1.Service, revisions []*revisionDesc) error {
	dw := printers.NewPrefixWriter(w)

	writeServiceCompact(dw, service)
	dw.WriteLine(separator())
	if err := dw.Flush(); err != nil {
		return err
	}

	writeRevisionsCompact(dw, revisions)
	dw.WriteLine(separator())
	if err := dw.Flush(); err != nil {
		return err
	}

	writeConditionsCompact(dw, service)
	dw.WriteLine(separator())
	if err := dw.Flush(); err != nil {
		return err
	}

	return nil
}

func separator() string {
	return strings.Repeat("─", 80)
}

// Emojii variation of label
func e(label string) string {
	return emojiMap[label]
}

// Write out service compactly. We can't use a tabbed writer here as it can't be used runes (emojis)
func writeServiceCompact(dw printers.PrefixWriter, service *v1alpha1.Service) {
	dw.WriteLine(separator())
	dw.Write(printers.LEVEL_0, "%s %s │ %s %s │ %s %s\n",
		e("Name"), service.Name, e("Namespace"),
		service.Namespace, e("Age"), age(service.CreationTimestamp.Time))
	dw.WriteLine(separator())
	dw.Write(printers.LEVEL_0, "%s %s\n", e("URL"), extractURL(service))
	if service.Status.Address != nil {
		dw.Write(printers.LEVEL_0, "%s %s\n", e("Address"), service.Status.Address.Hostname)
	}
	writeMapDescCompact(dw, printers.LEVEL_0, service.Labels, e("Labels"))
	writeMapDescCompact(dw, printers.LEVEL_0, service.Annotations, e("Annotations"))
}

func writeRevisionsCompact(dw printers.PrefixWriter, revisions []*revisionDesc) {
	firstLabel := e("Revisions")
	for _, desc := range revisions {
		dw.Write(printers.LEVEL_0, "%s %4s %s %s\n", firstLabel, getPercentageCompact(desc), e("Name"), getRevisionNameWithGenerationAndAge(desc))
		dw.Write(printers.LEVEL_0, "        %s %s\n", e("Image"), getImageDesc(desc))
		if desc.port != nil {
			dw.WriteColsLn(printers.LEVEL_0, "      %s %s\n", e("Port"), strconv.FormatInt(int64(*desc.port), 10))
		}
		writeEnvCompact(dw, desc.env)
		firstLabel = " "
	}
}

func writeConditionsCompact(dw printers.PrefixWriter, service *v1alpha1.Service) {
	maxLen := getMaxTypeLen(service.Status.Conditions)
	formatRow := "%s   %s %-" + strconv.Itoa(maxLen+colorOffset()) + "s %6s %-s\n"
	firstLabel := extractConditionsEmoji(service)
	for _, condition := range service.Status.Conditions {
		status := formatConditionStatusEmoji(condition.Status)
		conditionType := formatConditionType(condition)
		reason := condition.Reason
		age := wc(age(condition.LastTransitionTime.Inner.Time), color.FgHiBlack)
		if printAll && reason != "" {
			reason = fmt.Sprintf("%s (%s)", reason, condition.Message)
		}
		dw.Write(printers.LEVEL_0, formatRow, firstLabel, status, conditionType, age, reason)
		firstLabel = " "
	}
}

func formatConditionStatusEmoji(status v1.ConditionStatus) string {
	switch status {
	case v1.ConditionTrue:
		return "👍"
	case v1.ConditionFalse:
		return "👎"
	}
	return "🤔"

}

func extractConditionsEmoji(service *v1alpha1.Service) string {
	ret := "☀️"
	for _, condition := range service.Status.Conditions {
		if condition.Status == v1.ConditionFalse {
			return "🌧️"
		}
		if condition.Status == v1.ConditionUnknown {
			ret = "☁️"
		}
	}
	return ret
}

func getPercentageCompact(desc *revisionDesc) string {
	percent := desc.percent
	if percent == 100 {
		return "💯 "
	}
	if percent == 0 {
		return "💤 "
	}
	return fmt.Sprintf("%4d%%", desc.percent)
}

// ==================================

func writeMapDescCompact(dw printers.PrefixWriter, indent int, data map[string]string, label string) {
	if len(data) == 0 {
		return
	}

	if printAll {
		prefix := label
		for key, value := range data {
			dw.Write(indent, "%s  %s=%s\n", prefix, key, value)
			prefix = " "
		}
		return
	}

	dw.Write(indent, "%s  %s\n", label, joinAndTruncate(data))
}

func writeEnvCompact(dw printers.PrefixWriter, env []string) {
	if len(env) == 0 {
		return
	}
	label := e("Env")
	if printAll {
		prefix := label
		for _, value := range env {
			dw.Write(printers.LEVEL_0, "        %s %s\n", prefix, value)
			prefix = "  "
		}
		return
	}

	joined := strings.Join(env, ", ")
	if len(joined) > truncateAt {
		joined = joined[:truncateAt-4] + " ..."
	}
	dw.Write(printers.LEVEL_0, "        %s %s\n", label, joined)
}
