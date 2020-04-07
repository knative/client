/*
Copyright 2019 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// The following is a subset of original implementation
// at https://github.com/kubernetes/kubernetes/blob/v1.15.0-alpha.2/pkg/printers/tableprinter.go

package printers

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"text/tabwriter"

	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ ResourcePrinter = &HumanReadablePrinter{}

// NewTablePrinter creates a printer suitable for calling PrintObj().
func NewTablePrinter(options PrintOptions) *HumanReadablePrinter {
	printer := &HumanReadablePrinter{
		handlerMap: make(map[reflect.Type]*handlerEntry),
		options:    options,
	}
	return printer
}

// PrintObj prints the obj in a human-friendly format according to the type of the obj.
func (h *HumanReadablePrinter) PrintObj(obj runtime.Object, output io.Writer) error {
	if obj == nil {
		return nil
	}

	w, found := output.(*tabwriter.Writer)
	if !found {
		w = NewTabWriter(output)
		output = w
		defer w.Flush()
	}

	// Search for a handler registered handler to print it
	t := reflect.TypeOf(obj)
	if handler := h.handlerMap[t]; handler != nil {

		if err := printRowsForHandlerEntry(output, handler, obj, h.options); err != nil {
			return err
		}
		return nil
	}

	// we failed all reasonable printing efforts, report failure
	return fmt.Errorf("error: unknown type %#v", obj)
}

// printRowsForHandlerEntry prints the incremental table output
// including all the rows in the object. It returns the current type
// or an error, if any.
func printRowsForHandlerEntry(output io.Writer, handler *handlerEntry, obj runtime.Object, options PrintOptions) error {
	var results []reflect.Value

	args := []reflect.Value{reflect.ValueOf(obj), reflect.ValueOf(options)}
	results = handler.printFunc.Call(args)
	if !results[1].IsNil() {
		return results[1].Interface().(error)
	}

	if !options.NoHeaders {
		var headers []string
		for _, column := range handler.columnDefinitions {
			if !options.AllNamespaces && column.Priority == 0 {
				continue
			}
			headers = append(headers, strings.ToUpper(column.Name))
		}
		printHeader(headers, output)
	}

	if results[1].IsNil() {
		rows := results[0].Interface().([]metav1beta1.TableRow)
		printRows(output, rows, options)
		return nil
	}
	return results[1].Interface().(error)
}

func printHeader(columnNames []string, w io.Writer) error {
	if _, err := fmt.Fprintf(w, "%s\n", strings.Join(columnNames, "\t")); err != nil {
		return err
	}
	return nil
}

// printRows writes the provided rows to output.
func printRows(output io.Writer, rows []metav1beta1.TableRow, options PrintOptions) {
	for _, row := range rows {
		for i, cell := range row.Cells {
			if i != 0 {
				fmt.Fprint(output, "\t")
			}
			fmt.Fprint(output, cell)
		}
		output.Write([]byte("\n"))
	}
}
