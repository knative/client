/*
Copyright 2026 The Knative Authors

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
package flags

import "testing"

// A mount option without a value used to index past the end of the split.
func TestParseMountOptionsWithoutValue(t *testing.T) {
	tests := []struct {
		in      string
		wantErr bool
	}{
		{"readonly=true", false},
		{"readonly=false", false},
		{"readonly", true},
		{"readonly=", false},
		{"bogus=1", true},
	}

	for _, tc := range tests {
		_, err := parseMountOptions(tc.in)
		if tc.wantErr && err == nil {
			t.Errorf("%q: expected an error, got none", tc.in)
		}
		if !tc.wantErr && err != nil {
			t.Errorf("%q: unexpected error: %v", tc.in, err)
		}
	}
}
