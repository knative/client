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

package printers

import (
	"github.com/knative/client/pkg/serving"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"io"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions/printers"
)

type wrappedGvkUpdatePrinter struct {
	printer printers.ResourcePrinter
}

// Wrapper printer for updating an object with GVK just before printing
func NewGvkUpdatePrinter(delegate printers.ResourcePrinter) printers.ResourcePrinter {
	return wrappedGvkUpdatePrinter{delegate}
}

func (gvkPrinter wrappedGvkUpdatePrinter) PrintObj(obj runtime.Object, writer io.Writer) error {
	// TODO: How to deal with multiple versions which needs to be supported soon ?
	if err := serving.UpdateGroupVersionKind(obj, v1alpha1.SchemeGroupVersion); err != nil {
		return err
	}
	return gvkPrinter.printer.PrintObj(obj, writer)
}