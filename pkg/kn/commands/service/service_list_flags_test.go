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

package service

import (
	"reflect"
	"testing"

	"github.com/spf13/cobra"
	"gotest.tools/assert"
)

func TestServiceListFlags(t *testing.T) {
	var cmd *cobra.Command

	t.Run("adds service list flag", func(t *testing.T) {
		serviceListFlags := NewServiceListFlags()

		cmd = &cobra.Command{}
		serviceListFlags.AddFlags(cmd)

		assert.Assert(t, serviceListFlags != nil)
		assert.Assert(t, cmd.Flags() != nil)

		allowMissingTemplateKeys, err := cmd.Flags().GetBool("allow-missing-template-keys")
		assert.NilError(t, err)
		assert.Assert(t, allowMissingTemplateKeys == true)

		actualFormats := serviceListFlags.AllowedFormats()
		expectedFormats := []string{"json", "yaml", "name", "go-template", "go-template-file", "template", "templatefile", "jsonpath", "jsonpath-file", "no-headers"}
		assert.Assert(t, reflect.DeepEqual(actualFormats, expectedFormats))
	})
}
