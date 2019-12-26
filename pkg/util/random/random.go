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

package random

import (
	"math/rand"
	"strings"
)

var charChoices = []string{
	"b", "c", "d", "f", "g", "h", "j", "k", "l", "m", "n", "p", "q", "r", "s", "t", "v", "w", "x",
	"y", "z",
}

func Random(l int) string {
	chars := make([]string, 0, l)
	for i := 0; i < l; i++ {
		chars = append(chars, charChoices[rand.Int()%len(charChoices)])
	}
	return strings.Join(chars, "")

}
