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
	"bytes"
	"os"
	"testing"

	"gotest.tools/v3/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/client/pkg/util"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func TestPrintObj(t *testing.T) {
	h := NewTablePrinter(PrintOptions{})
	assert.Assert(t, h != nil)
	assert.NilError(t, h.PrintObj(nil, nil))

	myksvc := &servingv1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: "serving.knative.dev/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "myksvc", Namespace: "default"},
	}
	assert.ErrorContains(t, h.PrintObj(myksvc, os.Stdout), "unknown type")
	h.TableHandler(columnDefs, validPrintFunc)
	var out bytes.Buffer
	assert.NilError(t, h.PrintObj(myksvc, &out))
	assert.Assert(t, util.ContainsAll(out.String(), "myksvc"))
	h = NewTablePrinter(PrintOptions{})
	h.TableHandler(columnDefs, validPrintFuncErrOutput)
	assert.Error(t, h.PrintObj(myksvc, os.Stdout), mockErrorString)
}
