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

package flags

import (
	"bytes"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"gotest.tools/v3/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/duration"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"knative.dev/client/pkg/printers"
	hprinters "knative.dev/client/pkg/printers"
	"knative.dev/client/pkg/util"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
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
	myksvc = &servingv1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: "serving.knative.dev/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "myksvc", Namespace: "default"},
	}
)

func TestListPrintFlagsFormats(t *testing.T) {
	flags := NewListPrintFlags(nil)
	formats := flags.AllowedFormats()
	expected := []string{"json", "yaml", "name", "go-template", "go-template-file", "template", "templatefile", "jsonpath", "jsonpath-as-json", "jsonpath-file", "no-headers"}
	assert.DeepEqual(t, formats, expected)
}

func TestListPrintFlags(t *testing.T) {
	var cmd *cobra.Command
	flags := NewListPrintFlags(func(h hprinters.PrintHandler) {})

	cmd = &cobra.Command{}
	flags.AddFlags(cmd)

	assert.Assert(t, flags != nil)
	assert.Assert(t, cmd.Flags() != nil)

	allowMissingTemplateKeys, err := cmd.Flags().GetBool("allow-missing-template-keys")
	assert.NilError(t, err)
	assert.Assert(t, allowMissingTemplateKeys == true)

	p, err := flags.ToPrinter()
	assert.NilError(t, err)
	_, ok := p.(hprinters.ResourcePrinter)
	assert.Check(t, ok == true)
}

func TestListPrintFlagsPrint(t *testing.T) {
	var cmd *cobra.Command
	flags := NewListPrintFlags(func(h hprinters.PrintHandler) {
		h.TableHandler(columnDefs, validPrintFunc)
	})

	cmd = &cobra.Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)
	flags.AddFlags(cmd)

	pr, err := flags.ToPrinter()
	assert.NilError(t, err)
	assert.Assert(t, pr != nil)

	err = flags.Print(nil, cmd.OutOrStdout())
	assert.NilError(t, err)
	flags.GenericPrintFlags = genericclioptions.NewPrintFlags("mockOperation")
	flags.GenericPrintFlags.OutputFlagSpecified = func() bool {
		return true
	}
	err = flags.Print(myksvc, cmd.OutOrStdout())
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(out.String(), "myksvc"))
}

func TestEnsureNamespaces(t *testing.T) {
	var cmd *cobra.Command
	flags := NewListPrintFlags(func(h hprinters.PrintHandler) {
		h.TableHandler(columnDefs, validPrintFunc)
	})

	cmd = &cobra.Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)
	flags.AddFlags(cmd)
	err := flags.Print(myksvc, cmd.OutOrStderr())
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(out.String(), "myksvc"))
	assert.Assert(t, util.ContainsNone(out.String(), "default"))
	flags.EnsureWithNamespace()
	err = flags.Print(myksvc, cmd.OutOrStderr())
	assert.Assert(t, util.ContainsAll(out.String(), "default"))
	assert.NilError(t, err)
}
