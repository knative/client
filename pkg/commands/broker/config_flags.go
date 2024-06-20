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
	// KReferenceMapping is mapping between the known config kinds to a basic
	// default KReference value
	KReferenceMapping = map[ConfigType]*duckv1.KReference{
		ConfigMapType: {Kind: "ConfigMap", APIVersion: "v1"},
		SecretType:    {Kind: "Secret", APIVersion: "v1"},
		RabbitMqType:  {Kind: "RabbitmqCluster", APIVersion: "rabbitmq.com/v1beta1"},
	}
	ConfigTypeMapping = map[string]ConfigType{
		"cm":              ConfigMapType,
		"configmap":       ConfigMapType,
		"sc":              SecretType,
		"secret":          SecretType,
		"rabbitmqcluster": RabbitMqType,
		"rabbitmq":        RabbitMqType,
		"rmq":             RabbitMqType,
	}
)

// ConfigFlags represents the broker config
type ConfigFlags struct {
	BrokerConfig string
}

// Add is used to add the broker config flag to a command
func (c *ConfigFlags) Add(cmd *cobra.Command) {
	cmd.Flags().StringVar(&c.BrokerConfig, "broker-config", "", "Reference to the broker configuration "+
		"For example, a pointer to a ConfigMap (cm:, configmap:), Secret(sc:, secret:), RabbitmqCluster(rmq:, rabbitmq: rabbitmqcluster:) etc. "+
		"It should be used in conjunction with --class flag. "+
		"The format for specifying the object is a colon separated string consisting of at most 4 slices:\n"+
		"Length 1: <object-name> (the object will be assumed to be ConfigMap with the same name)\n"+
		"Length 2: <kind>:<object-name> (the APIVersion will be determined for ConfigMap, Secret, and RabbitmqCluster types)\n"+
		"Length 3: <kind>:<object-name>:<namespace> (the APIVersion will be determined only for ConfigMap, Secret, "+
		"and RabbitmqCluster types. Otherwise it will be interpreted as:\n"+
		"<apiVersion>:<kind>:<object-name>)\n"+
		"Length 4: <apiVersion>:<kind>:<object-name>:<namespace>")
}

// GetBrokerConfigReference parses the broker config
// and return the appropriate KReference object
func (c *ConfigFlags) GetBrokerConfigReference() (*duckv1.KReference, error) {
	config := c.BrokerConfig
	slices := strings.SplitN(config, ":", 4)
	if len(slices) == 1 {
		// If no APIVersion or Kind is specified, assume Configmap
		return &duckv1.KReference{
			Kind:       "ConfigMap",
			Name:       slices[0],
			APIVersion: "v1",
		}, nil
	} else if len(slices) == 2 {
		// If only two slices are present, it should resolve to
		// kind:name
		kind := slices[0]
		name := slices[1]
		kRef := getDefaultKReference(kind)
		if kRef.APIVersion == "" {
			return nil, fmt.Errorf("APIVersion could not be determined for kind %q. Provide config in format: \"<apiVersion>:<kind>:<name>:<namespace>\"", kind)
		}
		kRef.Name = name
		return kRef, nil
	} else if len(slices) == 3 {
		// 3 slices could resolve to either of the following:
		// 1. <kind>:<name>:<namespace>
		// 2. <apiVersion>:<kind>:<name>

		var kRef *duckv1.KReference
		if kRef = getDefaultKReference(slices[0]); kRef.APIVersion == "" {
			return &duckv1.KReference{
				Kind:       slices[1],
				Name:       slices[2],
				APIVersion: slices[0],
			}, nil
		}

		kRef.Name = slices[1]
		kRef.Namespace = slices[2]
		return kRef, nil
	} else {
		// 4 slices should resolve into <apiVersion>:<kind>:<name>:<namespace>
		return &duckv1.KReference{
			APIVersion: slices[0],
			Kind:       slices[1],
			Name:       slices[2],
			Namespace:  slices[3],
		}, nil
	}
}

func getDefaultKReference(kind string) *duckv1.KReference {
	k := strings.ToLower(kind)
	if configType, ok := ConfigTypeMapping[k]; ok {
		return KReferenceMapping[configType]
	}
	return &duckv1.KReference{Kind: kind}
}
