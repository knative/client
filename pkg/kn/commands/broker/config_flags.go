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
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"knative.dev/client/pkg/util"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

type ConfigType int

// Known config types for broker
const (
	ConfigMapType ConfigType = iota
	SecretType
	RabbitMqType
)

var (
	KReferenceMapping = map[ConfigType]*duckv1.KReference{
		ConfigMapType: {Kind: "ConfigMap", APIVersion: "v1"},
		SecretType:    {Kind: "Secret", APIVersion: "v1"},
		RabbitMqType:  {Kind: "RabbitmqCluster", APIVersion: "rabbitmq.com/v1beta1"},
	}
)

type ConfigFlags struct {
	BrokerConfig string
}

func (c *ConfigFlags) Add(cmd *cobra.Command) {
	cmd.Flags().StringVar(&c.BrokerConfig, "broker-config", "", "Broker config object like ConfigMap or RabbitMQ")
}

func (c *ConfigFlags) GetBrokerConfigReference() (*duckv1.KReference, error) {
	config := c.BrokerConfig
	slices := strings.SplitN(config, ":", 3)
	if len(slices) == 1 {
		return &duckv1.KReference{
			Kind:       "ConfigMap",
			Name:       slices[0],
			APIVersion: "v1",
		}, nil
	} else if len(slices) == 2 {
		kind := slices[0]
		name := slices[1]
		kRef := getDefaultKReference(kind)
		if kRef.APIVersion == "" {
			return nil, fmt.Errorf("kind %q is unknown and APIVersion could not be determined", kind)
		}
		kRef.Name = name
		return kRef, nil
	} else {
		kind := slices[0]
		name := slices[1]
		kRef := getDefaultKReference(kind)
		kRef.Name = name

		params := strings.Split(slices[2], ",")
		if len(params) == 1 && !strings.Contains(params[0], "=") {
			kRef.Namespace = params[0]
			return kRef, nil
		}
		mappedOptions, err := util.MapFromArray(params, "=")
		if err != nil {
			return nil, err
		}
		for k, v := range mappedOptions {
			switch strings.ToLower(k) {
			case "namespace":
				kRef.Namespace = v
			case "group":
				kRef.Group = v
			case "apiversion":
				kRef.APIVersion = v
			default:
				return nil, fmt.Errorf("incorrect field %q. Please specify any of the following: Namespace, Group, APIVersion", k)
			}
		}
		if kRef.APIVersion == "" {
			return nil, fmt.Errorf("kind %q is unknown and APIVersion could not be determined", kind)
		}
		return kRef, nil
	}
}

func getDefaultKReference(kind string) *duckv1.KReference {
	k := strings.ToLower(kind)
	switch k {
	case "cm", "configmap":
		return KReferenceMapping[ConfigMapType]
	case "sc", "secret":
		return KReferenceMapping[SecretType]
	case "rmq", "rabbitmq", "rabbitmqcluster":
		return KReferenceMapping[RabbitMqType]
	default:
		return &duckv1.KReference{Kind: kind}
	}
}
