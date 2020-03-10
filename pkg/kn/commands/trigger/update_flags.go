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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"knative.dev/client/pkg/util"
)

// TriggerUpdateFlags are flags for create and update a trigger
type TriggerUpdateFlags struct {
	Broker       string
	InjectBroker bool
	Filters      []string
}

// GetFilters to return a map type of filters
func (f *TriggerUpdateFlags) GetFilters() (map[string]string, error) {
	filters, err := util.MapFromArray(f.Filters, "=")
	if err != nil {
		return nil, errors.Wrap(err, "Invalid --filter")
	}
	return filters, nil
}

// GetUpdateFilters to return a map type of filters
func (f *TriggerUpdateFlags) GetUpdateFilters() (map[string]string, []string, error) {
	filters, err := util.MapFromArrayAllowingSingles(f.Filters, "=")
	if err != nil {
		return nil, nil, errors.Wrap(err, "Invalid --filter")
	}
	removes := util.ParseMinusSuffix(filters)
	return filters, removes, nil
}

//Add is to set parameters
func (f *TriggerUpdateFlags) Add(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.Broker, "broker", "default", "Name of the Broker which the trigger associates with.")
	cmd.Flags().BoolVar(&f.InjectBroker, "inject-broker", false, "Create new broker with name default through common annotation")
	cmd.Flags().StringSliceVar(&f.Filters, "filter", nil, "Key-value pair for exact CloudEvent attribute matching against incoming events, e.g type=dev.knative.foo")
}
