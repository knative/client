// Copyright 2020 The Knative Authors

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

const (
	FileModeReadWrite  = 0666
	FileModeExecutable = 0777
)

// GetResourceFieldsWithJSONPath returns output of given JSON path for given resource using kubectl and error if any
func GetResourceFieldsWithJSONPath(t *testing.T, it *KnTest, resource, name, jsonpath string) (string, error) {
	out, err := NewKubectl(it.Kn().Namespace()).Run("get", resource, name, "-o", jsonpath, "-n", it.Kn().Namespace())
	if err != nil {
		return "", err
	}

	return out, nil
}

// CreateFile creates a file with given name, content, path, fileMode and returns absolute filepath and error if any
func CreateFile(fileName, fileContent, filePath string, fileMode os.FileMode) (string, error) {
	file := filepath.Join(filePath, fileName)
	err := ioutil.WriteFile(file, []byte(fileContent), fileMode)
	return file, err
}
