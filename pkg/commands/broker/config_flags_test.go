// Copyright © 2022 The Knative Authors
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

package broker

import (
	"testing"

	"github.com/spf13/cobra"
	"gotest.tools/v3/assert"
	v1 "knative.dev/pkg/apis/duck/v1"
)

func TestConfigFlags_Add(t *testing.T) {
	testCmd := &cobra.Command{}
	c := ConfigFlags{}
	c.Add(testCmd)

	assert.NilError(t, testCmd.Flags().Set("broker-config", "mock-config"))
	assert.Equal(t, c.BrokerConfig, "mock-config")
}

func TestConfigFlags_GetBrokerConfigReference(t *testing.T) {
	tests := []struct {
		name          string
		argument      string
		expectedKRef  *v1.KReference
		expectedError string
	}{{
		name:     "no kind specified",
		argument: "mock-name",
		expectedKRef: &v1.KReference{
			Kind:       "ConfigMap",
			Namespace:  "",
			Name:       "mock-name",
			APIVersion: "v1",
			Group:      "",
		},
	},
		{
			name:     "only configmap kind and name specified",
			argument: "cm:mock-name",
			expectedKRef: &v1.KReference{
				Kind:       "ConfigMap",
				Namespace:  "",
				Name:       "mock-name",
				APIVersion: "v1",
				Group:      "",
			},
		},
		{
			name:     "only rabbitmq kind and name specified",
			argument: "rabbitmqcluster:mock-name",
			expectedKRef: &v1.KReference{
				Kind:       "RabbitmqCluster",
				Namespace:  "",
				Name:       "mock-name",
				APIVersion: "rabbitmq.com/v1beta1",
				Group:      "",
			},
		}, {
			name:          "only kind (unknown) and name specified without specifying API version",
			argument:      "unknown:mock-name",
			expectedKRef:  nil,
			expectedError: "APIVersion could not be determined for kind \"unknown\"",
		},
		{
			name:     "kind, name, and namespace specified",
			argument: "secret:mock-name:mock-namespace",
			expectedKRef: &v1.KReference{
				Kind:       "Secret",
				Namespace:  "mock-namespace",
				Name:       "mock-name",
				APIVersion: "v1",
				Group:      "",
			},
		},
		{
			name:     "apiVersion, kind, name, and namespace",
			argument: "rabbitmq.com/v1beta1:RabbitmqCluster:test-cluster:test-ns",
			expectedKRef: &v1.KReference{
				Kind:       "RabbitmqCluster",
				Namespace:  "test-ns",
				Name:       "test-cluster",
				APIVersion: "rabbitmq.com/v1beta1",
			},
		},
	}
	for _, tt := range tests {
		c := ConfigFlags{BrokerConfig: tt.argument}
		actualKRef, actualErr := c.GetBrokerConfigReference()
		assert.DeepEqual(t, tt.expectedKRef, actualKRef)
		if tt.expectedError == "" {
			assert.NilError(t, actualErr)
		} else {
			assert.ErrorContains(t, actualErr, tt.expectedError)
		}
	}
}
