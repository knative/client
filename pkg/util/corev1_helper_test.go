/*
Copyright 2020 The Knative Authors

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

package util

import (
	"testing"

	"gotest.tools/assert"
)

func TestGenerateVolumeName(t *testing.T) {
	actual := []string{
		"Ab12~`!@#$%^&*()-=_+[]{}|/\\<>,./?:;\"'xZ",
		"/Ab12~`!@#$%^&*()-=_+[]{}|/\\<>,./?:;\"'xZ/",
		"",
		"/",
		"/path.mypath/",
		"/.path.mypath",
	}

	expected := []string{
		"ab12---------------------------------xz",
		"ab12---------------------------------xz-",
		"k-",
		"k-",
		"path-mypath-",
		"k--path-mypath",
	}

	for i := range actual {
		actualName := GenerateVolumeName(actual[i])
		expectedName := appendCheckSum(expected[i], actual[i])
		assert.Equal(t, actualName, expectedName)
	}

	// 63 char limit case, no need to append the checksum in expected string
	expName_63 := "k---ab12---------------------------------xz-ab12--------------n"
	assert.Equal(t, len(expName_63), 63)
	assert.Equal(t, GenerateVolumeName("/./Ab12~`!@#$%^&*()-=_+[]{}|/\\<>,./?:;\"'xZ/Ab12~`!@#$%^&*()-=_+[]{}|/\\<>,./?:;\"'xZ/"), expName_63)
}
