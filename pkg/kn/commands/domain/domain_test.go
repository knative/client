// Copyright © 2021 The Knative Authors
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

package domain

import (
	"bytes"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"

	kndynamic "knative.dev/client/pkg/dynamic"
	"knative.dev/client/pkg/kn/commands"
	knflags "knative.dev/client/pkg/kn/flags"
	clientservingv1alpha1 "knative.dev/client/pkg/serving/v1alpha1"
	v1 "knative.dev/pkg/apis/duck/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	servingv1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"
)

// Helper methods
var blankConfig clientcmd.ClientConfig

const kubeConfig = `kind: Config
version: v1
users:
- name: u
clusters:
- name: c
  cluster:
    server: example.com
contexts:
- name: x
  context:
    user: u
    cluster: c
current-context: x`

func init() {
	var err error
	blankConfig, err = clientcmd.NewClientConfigFromBytes([]byte(kubeConfig))
	if err != nil {
		panic(err)
	}
}

func TestDomainCommand(t *testing.T) {
	knParams := &commands.KnParams{}
	domainCmd := NewDomainCommand(knParams)
	assert.Equal(t, domainCmd.Name(), "domain")
	assert.Equal(t, domainCmd.Use, "domain COMMAND")
	subCommands := make([]string, 0, len(domainCmd.Commands()))
	for _, cmd := range domainCmd.Commands() {
		subCommands = append(subCommands, cmd.Name())
	}
	expectedSubCommands := []string{"create", "delete", "describe", "list", "update"}
	assert.DeepEqual(t, subCommands, expectedSubCommands)
}

func executeDomainCommand(client clientservingv1alpha1.KnServingClient, dynamicClient kndynamic.KnDynamicClient, args ...string) (string, error) {
	knParams := &commands.KnParams{}
	knParams.ClientConfig = blankConfig

	output := new(bytes.Buffer)
	knParams.Output = output
	knParams.NewServingV1alpha1Client = func(namespace string) (clientservingv1alpha1.KnServingClient, error) {
		return client, nil
	}
	knParams.NewDynamicClient = func(namespace string) (kndynamic.KnDynamicClient, error) {
		return dynamicClient, nil
	}

	cmd := NewDomainCommand(knParams)
	cmd.SetArgs(args)
	cmd.SetOut(output)

	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return knflags.ReconcileBoolFlags(cmd.Flags())
	}
	err := cmd.Execute()
	return output.String(), err
}

func createService(name string) *servingv1.Service {
	return &servingv1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default"},
	}
}

func createDomainMapping(name string, ref v1.KReference) *servingv1alpha1.DomainMapping {
	return clientservingv1alpha1.NewDomainMappingBuilder(name).Namespace("default").Reference(ref).Build()
}

func createServiceRef(service, namespace string) v1.KReference {
	return v1.KReference{Name: service,
		Kind:       "Service",
		APIVersion: "serving.knative.dev/v1",
		Namespace:  namespace,
	}
}
