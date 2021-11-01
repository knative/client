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

package errors

import (
	"testing"

	"knative.dev/client/pkg/util"

	"gotest.tools/v3/assert"
)

func TestNewInvalidCRD(t *testing.T) {
	err := NewInvalidCRD("serving.knative.dev")
	assert.Assert(t, util.ContainsAll(err.Error(), "no or newer Knative Serving API found on the backend", "please verify the installation", "update", "'kn'"))

	err = NewInvalidCRD("eventing")
	assert.Assert(t, util.ContainsAll(err.Error(), "no or newer Knative Eventing API found on the backend", "please verify the installation", "update", "'kn'"))

	err = NewInvalidCRD("")
	assert.Assert(t, util.ContainsAll(err.Error(), "no or newer Knative  API found on the backend", "please verify the installation", "update", "'kn'"))
}
