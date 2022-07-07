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
			expectedError: "kind \"unknown\" is unknown and APIVersion could not be determined",
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
			name:     "kind, name, and namespace specified(key = val format)",
			argument: "secret:mock-name:Namespace=mock-namespace",
			expectedKRef: &v1.KReference{
				Kind:       "Secret",
				Namespace:  "mock-namespace",
				Name:       "mock-name",
				APIVersion: "v1",
				Group:      "",
			},
		},
		{
			name:     "kind, name, and namespace and group specified(key = val format)",
			argument: "rabbitmq:mock-name:Namespace=mock-namespace,group=rabbitmq.com",
			expectedKRef: &v1.KReference{
				Kind:       "RabbitmqCluster",
				Namespace:  "mock-namespace",
				Name:       "mock-name",
				APIVersion: "rabbitmq.com/v1beta1",
				Group:      "rabbitmq.com",
			},
		},
		{
			name:     "unknown kind, name, and namespace and APIVersion specified(key = val format)",
			argument: "unknown:mock-name:Namespace=mock-namespace,apiVersion=v1beta1",
			expectedKRef: &v1.KReference{
				Kind:       "unknown",
				Namespace:  "mock-namespace",
				Name:       "mock-name",
				APIVersion: "v1beta1",
				Group:      "",
			},
		},
		{
			name:          "unknown kind specified without APIVersion (key = val format)",
			argument:      "unknown:mock-name:Namespace=mock-namespace",
			expectedKRef:  nil,
			expectedError: "kind \"unknown\" is unknown and APIVersion could not be determined",
		},
		{
			name:          "kind, name, and an unknown key specified in third part of the argument",
			argument:      "secret:mock-name:unknown-key=mock-val",
			expectedKRef:  nil,
			expectedError: "incorrect field \"unknown-key\". Please specify any of the following: Namespace, Group, APIVersion",
		}}
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
