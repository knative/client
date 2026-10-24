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
