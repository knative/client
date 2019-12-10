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
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or im
// See the License for the specific language governing permissions and
// limitations under the License.

package source

import (
	"testing"

	hprinters "knative.dev/client/pkg/printers"

	"gotest.tools/assert"
)

func TestListTypesFlagsFormats(t *testing.T) {
	flags := NewListTypesFlags()
	formats := flags.AllowedFormats()
	expected := []string{"json", "yaml", "name", "go-template", "go-template-file", "template", "templatefile", "jsonpath", "jsonpath-file", "no-headers"}
	assert.DeepEqual(t, formats, expected)
}

func TestListTypesFlagsPrinter(t *testing.T) {
	flags := NewListTypesFlags()
	p, err := flags.GenericPrintFlags.ToPrinter()
	assert.NilError(t, err)
	_, ok := p.(hprinters.ResourcePrinter)
	assert.Check(t, ok == true)
}
