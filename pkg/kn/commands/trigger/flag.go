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

package trigger

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type filterArray []string

func (filters *filterArray) String() string {
	str := ""
	for _, filter := range *filters {
		str = str + filter + " "
	}
	return str
}

func (filters *filterArray) Set(value string) error {
	*filters = append(*filters, value)
	return nil
}

func (filters *filterArray) Type() string {
	return "[]string"
}

// TriggerUpdateFlags are flags for create and update a trigger
type TriggerUpdateFlags struct {
	Broker  string
	Filters filterArray
}

// GetFilter to return a map type of filters
func (f *TriggerUpdateFlags) GetFilters() map[string]string {
	filters := map[string]string{}
	for _, item := range f.Filters {
		parts := strings.Split(item, "=")
		if len(parts) == 2 {
			filters[parts[0]] = parts[1]
		} else {
			fmt.Printf("Ignore invalid filter %s", f)
		}
	}
	return filters
}

//Add is to set parameters
func (f *TriggerUpdateFlags) Add(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.Broker, "broker", "default", "Name of the Broker which the trigger associates with.")
	cmd.Flags().Var(&f.Filters, "filter", "Key-value pair for exact CloudEvent attribute matching against incoming events, e.g type=dev.knative.foo")
}
