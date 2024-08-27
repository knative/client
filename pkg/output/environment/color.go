/*
 Copyright 2024 The Knative Authors

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

package environment

import (
	"os"
	"strings"
)

// SupportsColor returns true if the environment supports color output.
//
// See NonColorRequested and ColorIsForced functions for more information.
func SupportsColor() bool {
	color := true
	if NonColorRequested() {
		color = false
	}
	if ColorIsForced() {
		color = true
	}
	return color
}

// NonColorRequested returns true if the NO_COLOR environment variable is set to a
// truthy value.
//
// See https://no-color.org/ for more information.
func NonColorRequested() bool {
	return settingToBool(os.Getenv("NO_COLOR"))
}

// ColorIsForced returns true if the FORCE_COLOR environment variable is set to
// a truthy value.
//
// See https://force-color.org/ for more information.
func ColorIsForced() bool {
	return settingToBool(os.Getenv("FORCE_COLOR"))
}

func settingToBool(s string) bool {
	s = strings.ToLower(s)
	return len(s) != 0 &&
		s != "0" &&
		s != "false" &&
		s != "no" &&
		s != "off"
}
