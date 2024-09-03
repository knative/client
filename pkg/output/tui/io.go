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

package tui

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"knative.dev/client/pkg/output"
)

func ioProgramOptions(io output.InputOutput) []tea.ProgramOption {
	opts := make([]tea.ProgramOption, 0, 2)
	opts = append(opts, tea.WithInput(safeguardBubbletea964(io.InOrStdin())))
	if io.OutOrStdout() != nil && io.OutOrStdout() != os.Stdout {
		opts = append(opts, tea.WithOutput(io.OutOrStdout()))
	}
	return opts
}
