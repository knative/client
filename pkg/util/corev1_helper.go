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
	"crypto/sha1"
	"fmt"
	"strings"
	"unicode"

	corev1 "k8s.io/api/core/v1"
)

// EnvToMap is an utility function to translate between the API list form of env vars, and the
// more convenient map form.
func EnvToMap(vars []corev1.EnvVar) (map[string]string, error) {
	result := map[string]string{}
	for _, envVar := range vars {
		_, present := result[envVar.Name]
		if present {
			return nil, fmt.Errorf("env var name present more than once: %v", envVar.Name)
		}
		result[envVar.Name] = envVar.Value
	}
	return result, nil
}

// GenerateVolumeName generates a volume name with respect to a given path string.
// Current implementation basically sanitizes the path string by replacing "/" with "-"
// To reduce any chance of duplication, a checksum part generated from the path string is appended to the sanitized string.
// The volume name must follow the DNS label standard as defined in RFC 1123. This means the name must:
// - contain at most 63 characters
// - contain only lowercase alphanumeric characters or '-'
// - start with an alphanumeric character
// - end with an alphanumeric character
func GenerateVolumeName(path string) string {
	builder := &strings.Builder{}
	for idx, r := range path {
		switch {
		case unicode.IsLower(r) || unicode.IsDigit(r) || r == '-':
			builder.WriteRune(r)
		case unicode.IsUpper(r):
			builder.WriteRune(unicode.ToLower(r))
		case r == '/':
			if idx != 0 {
				builder.WriteRune('-')
			}
		default:
			builder.WriteRune('-')
		}
	}

	vname := appendCheckSum(builder.String(), path)

	// the name must start with an alphanumeric character
	if !unicode.IsLetter(rune(vname[0])) && !unicode.IsNumber(rune(vname[0])) {
		vname = "k-" + vname
	}

	// contain at most 63 characters
	if len(vname) > 63 {
		// must end with an alphanumeric character
		vname = fmt.Sprintf("%s-n", vname[0:61])
	}

	return vname
}

func appendCheckSum(sanitizedString, path string) string {
	checkSum := sha1.Sum([]byte(path))
	shortCheckSum := checkSum[0:4]
	return fmt.Sprintf("%s-%x", sanitizedString, shortCheckSum)
}
