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

	"gotest.tools/assert"
)

func TestNewKNError(t *testing.T) {
	err := NewKNError("myerror")
	assert.Error(t, err, "myerror")

	err = NewKNError("")
	assert.Error(t, err, "")
}

func TestKNError_Error(t *testing.T) {
	err := NewKNError("myerror")
	assert.Equal(t, err.Error(), "myerror")

	err = NewKNError("")
	assert.Equal(t, err.Error(), "")
}
