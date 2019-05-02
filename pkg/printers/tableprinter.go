package printers

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"text/tabwriter"
)

var _ ResourcePrinter = &HumanReadablePrinter{}

var withNamespacePrefixColumns = []string{"NAMESPACE"}

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
	w, found := output.(*tabwriter.Writer)
	if !found {
		w = GetNewTabWriter(output)
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

	var headers []string
	for _, column := range handler.columnDefinitions {
		headers = append(headers, strings.ToUpper(column.Name))
	}
	printHeader(headers, output)

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
		if options.WithNamespace {
			if obj := row.Object.Object; obj != nil {
				if m, err := meta.Accessor(obj); err == nil {
					fmt.Fprint(output, m.GetNamespace())
				}
			}
			fmt.Fprint(output, "\t")
		}

		for i, cell := range row.Cells {
			if i != 0 {
				fmt.Fprint(output, "\t")
			} else {
				if options.WithKind && !options.Kind.Empty() {
					fmt.Fprintf(output, "%s/%s", strings.ToLower(options.Kind.String()), cell)
					continue
				}
			}
			fmt.Fprint(output, cell)
		}

		output.Write([]byte("\n"))
	}
}
