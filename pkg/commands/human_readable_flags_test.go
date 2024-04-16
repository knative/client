// Copyright © 2021 The Knative Authors
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
	"testing"
	"time"

	"gotest.tools/v3/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/duration"
	"knative.dev/client-pkg/pkg/printers"
	"knative.dev/client-pkg/pkg/util"
)

var (
	validPrintFunc = func(obj *servingv1.Service, opts printers.PrintOptions) ([]metav1beta1.TableRow, error) {
		tableRow := metav1beta1.TableRow{
			Object: runtime.RawExtension{Object: obj},
		}
		tableRow.Cells = append(tableRow.Cells, obj.Name, duration.HumanDuration(time.Since(obj.CreationTimestamp.Time)))
		if opts.AllNamespaces {
			tableRow.Cells = append(tableRow.Cells, obj.Namespace)
		}
		return []metav1.TableRow{tableRow}, nil
	}
	columnDefs = []metav1beta1.TableColumnDefinition{
		{Name: "Namespace", Type: "string", Description: "Namespace of mock instance", Priority: 0},
		{Name: "Name", Type: "string", Description: "Name of the mock instance", Priority: 1},
		{Name: "Age", Type: "string", Description: "Age of the mock instance", Priority: 1},
	}
)

func TestToPrinter(t *testing.T) {
	h := NewHumanPrintFlags()
	hprinter, err := h.ToPrinter(func(p printers.PrintHandler) {
		p.TableHandler(columnDefs, validPrintFunc)
	})
	assert.NilError(t, err)
	myksvc := &servingv1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: "serving.knative.dev/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "myksvc", Namespace: "default"},
	}
	var out bytes.Buffer
	assert.NilError(t, hprinter.PrintObj(myksvc, &out))
	assert.Assert(t, util.ContainsAll(out.String(), "myksvc"))
	assert.Assert(t, util.ContainsNone(out.String(), "default"))

	h.EnsureWithNamespace()
	hprinter, err = h.ToPrinter(func(p printers.PrintHandler) {
		p.TableHandler(columnDefs, validPrintFunc)
	})
	assert.NilError(t, err)
	assert.NilError(t, hprinter.PrintObj(myksvc, &out))
	assert.Assert(t, util.ContainsAll(out.String(), "myksvc"))
	assert.Assert(t, util.ContainsAll(out.String(), "default"))
}

func TestConditionsValue(t *testing.T) {
	conds := duckv1.Conditions([]apis.Condition{{Status: corev1.ConditionTrue}, {Status: corev1.ConditionTrue}, {Status: corev1.ConditionFalse}, {Status: corev1.ConditionUnknown}})
	expected := "2 OK / 4"
	assert.Equal(t, expected, ConditionsValue(conds))
}

func TestReadyCondtion(t *testing.T) {
	conds := duckv1.Conditions([]apis.Condition{{Status: corev1.ConditionTrue}, {Status: corev1.ConditionTrue}, {Status: corev1.ConditionFalse, Type: apis.ConditionReady}, {Status: corev1.ConditionUnknown, Type: apis.ConditionReady}})
	expected := string(corev1.ConditionFalse)
	assert.Equal(t, expected, ReadyCondition(conds))
	assert.Equal(t, "<unknown>", ReadyCondition(conds[:1]))
}

func TestNonReadyConditionReason(t *testing.T) {
	conds := duckv1.Conditions([]apis.Condition{{Type: apis.ConditionReady, Status: corev1.ConditionTrue}, {Type: apis.ConditionReady, Message: "mockMsg", Reason: "mockReason", Status: corev1.ConditionFalse}, {Reason: "mockReason", Status: corev1.ConditionFalse, Type: apis.ConditionReady}, {Status: corev1.ConditionUnknown, Type: apis.ConditionReady}})
	assert.Equal(t, "", NonReadyConditionReason(conds))
	expected := "mockReason : mockMsg"
	assert.Equal(t, expected, NonReadyConditionReason(conds[1:]))
	assert.Equal(t, "mockReason", NonReadyConditionReason(conds[2:]))

	nonReadyConds := duckv1.Conditions([]apis.Condition{{Type: apis.ConditionSucceeded}, {Type: apis.ConditionType(apis.ConditionSeverityError)}, {Type: apis.ConditionType(apis.ConditionSeverityInfo)}})
	assert.Equal(t, "<unknown>", NonReadyConditionReason(nonReadyConds))
}

func TestTranslateTimestampSince(t *testing.T) {
	ts := time.Time{}
	assert.Equal(t, "<unknown>", TranslateTimestampSince(metav1.NewTime(ts)))
	assert.Equal(t, "0s", TranslateTimestampSince(metav1.Now()))
}
