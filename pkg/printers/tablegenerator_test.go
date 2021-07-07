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

package printers

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/duration"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"gotest.tools/v3/assert"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
)

const (
	mockErrorString = "This function returns an error"
)

var (
	invalidPrintTypeFunc             = true
	invalidPrintFuncIncorrectInCount = func() ([]metav1beta1.TableRow, error) { return nil, nil }
	invalidPrintFuncIncorrectInType  = func(*servingv1.Service, bool) ([]metav1beta1.TableRow, error) { return nil, nil }
	validPrintFunc                   = func(obj *servingv1.Service, opts PrintOptions) ([]metav1beta1.TableRow, error) {
		tableRow := metav1beta1.TableRow{
			Object: runtime.RawExtension{Object: obj},
		}
		tableRow.Cells = append(tableRow.Cells, obj.Namespace, obj.Name, duration.HumanDuration(time.Since(obj.CreationTimestamp.Time)))
		return []metav1.TableRow{tableRow}, nil
	}
	validPrintFuncErrOutput = func(obj *servingv1.Service, opts PrintOptions) ([]metav1beta1.TableRow, error) {
		return nil, fmt.Errorf(mockErrorString)
	}
	columnDefs = []metav1beta1.TableColumnDefinition{
		{Name: "Namespace", Type: "string", Description: "Namespace of mock instance", Priority: 0},
		{Name: "Name", Type: "string", Description: "Name of the mock instance", Priority: 1},
		{Name: "Age", Type: "string", Description: "Age of the mock instance", Priority: 1},
	}
)

// Do we need this function?
func TestValidateRowPrintHandlerFunc(t *testing.T) {
	assert.Error(t, ValidateRowPrintHandlerFunc(reflect.ValueOf(invalidPrintTypeFunc)), fmt.Sprintf("invalid print handler. %#v is not a function", reflect.ValueOf(invalidPrintTypeFunc)))
	assert.Error(t, ValidateRowPrintHandlerFunc(reflect.ValueOf(invalidPrintFuncIncorrectInCount)), fmt.Sprintf("invalid print handler."+
		"Must accept 2 parameters and return 2 value."))
	assert.Error(t, ValidateRowPrintHandlerFunc(reflect.ValueOf(invalidPrintFuncIncorrectInType)), fmt.Sprintf("invalid print handler. The expected signature is: "+
		"func handler(obj %v, options PrintOptions) ([]metav1beta1.TableRow, error)", reflect.TypeOf(&servingv1.Service{})))
	assert.NilError(t, ValidateRowPrintHandlerFunc(reflect.ValueOf(validPrintFunc)))
}

func TestTableHandler(t *testing.T) {
	h := NewTableGenerator()
	assert.ErrorContains(t, h.TableHandler(columnDefs, invalidPrintTypeFunc), "invalid print handler")
	assert.ErrorContains(t, h.TableHandler(columnDefs, invalidPrintFuncIncorrectInCount), "invalid print handler")
	assert.ErrorContains(t, h.TableHandler(columnDefs, invalidPrintFuncIncorrectInType), "invalid print handler")
	assert.NilError(t, h.TableHandler(columnDefs, validPrintFunc))
	assert.ErrorContains(t, h.TableHandler(columnDefs, validPrintFunc), "registered duplicate printer")
}

func TestGenerateTable(t *testing.T) {
	h := NewTableGenerator()
	myksvc := &servingv1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: "serving.knative.dev/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "myksvc", Namespace: "default"},
	}
	_, err := h.GenerateTable(myksvc, h.options)
	assert.ErrorContains(t, err, "no table handler registered for this type")
	err = h.TableHandler(columnDefs, validPrintFunc)
	assert.NilError(t, err)
	table, err := h.GenerateTable(myksvc, h.options)
	assert.NilError(t, err)
	expected := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: myksvc},
	}
	expected.Cells = append(expected.Cells, myksvc.Namespace, myksvc.Name, duration.HumanDuration(time.Since(myksvc.CreationTimestamp.Time)))
	assert.DeepEqual(t, table.Rows[0], expected)

	myksvcList := &servingv1.ServiceList{
		TypeMeta: metav1.TypeMeta{},
		ListMeta: metav1.ListMeta{},
		Items:    []servingv1.Service{*myksvc},
	}

	printServiceList := func(objList *servingv1.ServiceList, opts PrintOptions) ([]metav1beta1.TableRow, error) {
		rows := make([]metav1beta1.TableRow, 0, len(objList.Items))
		for _, obj := range objList.Items {
			row, err := validPrintFunc(&obj, h.options)
			assert.NilError(t, err)
			rows = append(rows, row...)
		}
		return rows, nil
	}
	err = h.TableHandler(columnDefs, printServiceList)
	assert.NilError(t, err)
	table, err = h.GenerateTable(myksvcList, h.options)
	assert.NilError(t, err)
	assert.DeepEqual(t, table.Rows[0], expected)

	// Create new human readable printer
	h = NewTableGenerator()
	assert.NilError(t, h.TableHandler(columnDefs, validPrintFuncErrOutput))
	_, err = h.GenerateTable(myksvc, h.options)
	assert.Error(t, err, mockErrorString)
}
