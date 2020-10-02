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

package printing

import (
	"fmt"

	"knative.dev/client/pkg/printers"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

// DescribeSink prints the given 'sink' for the given prefix writer 'dw',
// provide 'attribute' to print the section heading for this sink
func DescribeSink(dw printers.PrefixWriter, attribute, namespace string, sink *duckv1.Destination) {
	if sink == nil {
		return
	}
	subWriter := dw.WriteAttribute(attribute, "")
	ref := sink.Ref
	if ref != nil {
		subWriter.WriteAttribute("Name", sink.Ref.Name)
		if sink.Ref.Namespace != "" && sink.Ref.Namespace != namespace {
			subWriter.WriteAttribute("Namespace", sink.Ref.Namespace)
		}
		subWriter.WriteAttribute("Resource", fmt.Sprintf("%s (%s)", sink.Ref.Kind, sink.Ref.APIVersion))
	}
	uri := sink.URI
	if uri != nil {
		subWriter.WriteAttribute("URI", uri.String())
	}
}
