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
	"testing"

	"github.com/spf13/cobra"
	"gotest.tools/v3/assert"
	hprinters "knative.dev/client/pkg/printers"
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
	flags := NewListPrintFlags(func(h hprinters.PrintHandler) {})

	cmd = &cobra.Command{}
	flags.AddFlags(cmd)

	pr, err := flags.ToPrinter()
	assert.NilError(t, err)
	assert.Assert(t, pr != nil)

	err = flags.Print(nil, cmd.OutOrStdout())
	assert.NilError(t, err)
}
