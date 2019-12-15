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

package apiserver

import (
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"knative.dev/eventing/pkg/apis/sources/v1alpha1"
)

const (
	ApiVersionSplitChar = ":"
	DefaultApiVersion   = "v1"
)

// ApiServerSourceUpdateFlags are flags for create and update a ApiServerSource
type ApiServerSourceUpdateFlags struct {
	ServiceAccountName string
	Mode               string
	Resources          []string
}

// GetApiServerResourceArray is to return an array of ApiServerResource from a string. A sample is Event:v1:true,Pod:v2:false
func (f *ApiServerSourceUpdateFlags) GetApiServerResourceArray() []v1alpha1.ApiServerResource {
	var resourceList []v1alpha1.ApiServerResource
	for _, r := range f.Resources {
		version, kind, controller := getValidResource(r)
		resourceRef := v1alpha1.ApiServerResource{
			APIVersion: version,
			Kind:       kind,
			Controller: controller,
		}
		resourceList = append(resourceList, resourceRef)
	}
	return resourceList
}

func getValidResource(resource string) (string, string, bool) {
	var version = DefaultApiVersion // v1 as default
	var isController = false        //false as default
	var err error

	parts := strings.Split(resource, ApiVersionSplitChar)
	kind := parts[0]
	if len(parts) >= 2 {
		version = parts[1]
	}
	if len(parts) >= 3 {
		isController, err = strconv.ParseBool(parts[2])
		if err != nil {
			isController = false
		}
	}
	return version, kind, isController
}

//Add is to set parameters
func (f *ApiServerSourceUpdateFlags) Add(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.ServiceAccountName,
		"service-account",
		"",
		"Name of the service account to use to run this source")
	cmd.Flags().StringVar(&f.Mode,
		"mode",
		"Ref",
		`The mode the receive adapter controller runs under:, 
"Ref" sends only the reference to the resource, 
"Resource" send the full resource.`)
	cmd.Flags().StringSliceVar(&f.Resources,
		"resource",
		nil,
		`Comma seperate Kind:APIVersion:isController list, e.g. Event:v1:true.
"APIVersion" and "isControler" can be omitted.
"APIVersion" is "v1" by default, "isController" is "false" by default.`)
}

//Apply updates the service object based on flags
func (f *ApiServerSourceUpdateFlags) Apply(source *v1alpha1.ApiServerSource, cmd *cobra.Command) {
	if cmd.Flags().Changed("service-account") {
		source.Spec.ServiceAccountName = f.ServiceAccountName
	}

	if cmd.Flags().Changed("mod") {
		source.Spec.Mode = f.Mode
	}

	if cmd.Flags().Changed("resource") {
		source.Spec.Resources = f.GetApiServerResourceArray()
	}
}
