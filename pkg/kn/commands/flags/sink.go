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

package flags

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"

	"knative.dev/client/pkg/serving/v1alpha1"
)

type SinkFlags struct {
	sink string
}

func (i *SinkFlags) Add(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&i.sink, "sink", "s", "", "Addressable sink for events")
}

func (i *SinkFlags) ResolveSink(client v1alpha1.KnServingClient) (*duckv1beta1.Destination, error) {
	if i.sink == "" {
		return nil, nil
	}

	if strings.HasPrefix(i.sink, "svc:") {
		serviceName := i.sink[4:]
		service, err := client.GetService(serviceName)
		if err != nil {
			return nil, err
		}
		return &duckv1beta1.Destination{
			Ref: &v1.ObjectReference{
				Kind:       service.Kind,
				APIVersion: service.APIVersion,
				Name:       service.Name,
				Namespace:  service.Namespace,
			},
		}, nil
	}

	return nil, fmt.Errorf("Not supported sink type: %s", i.sink)
}
