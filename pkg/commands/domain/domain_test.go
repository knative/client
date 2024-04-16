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
	"context"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"

	kndynamic "knative.dev/client-pkg/pkg/dynamic"
	dynamicfake "knative.dev/client-pkg/pkg/dynamic/fake"
	clientservingv1beta1 "knative.dev/client-pkg/pkg/serving/v1beta1"
	"knative.dev/client/pkg/commands"
	knflags "knative.dev/client/pkg/flags"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	servingv1beta1 "knative.dev/serving/pkg/apis/serving/v1beta1"
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

type resolveCase struct {
	ref         string
	destination *duckv1.KReference
	errContents string
}

func TestResolve(t *testing.T) {
	myksvc := &servingv1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: "serving.knative.dev/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "myksvc", Namespace: "default"},
	}
	mykroute := &servingv1.Route{
		TypeMeta:   metav1.TypeMeta{Kind: "Route", APIVersion: "serving.knative.dev/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "mykroute", Namespace: "default"},
	}
	myksvcInOther := &servingv1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: "serving.knative.dev/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "myksvc", Namespace: "other"},
	}
	mykrouteInOther := &servingv1.Route{
		TypeMeta:   metav1.TypeMeta{Kind: "Route", APIVersion: "serving.knative.dev/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "mykroute", Namespace: "other"},
	}

	cases := []resolveCase{
		// Test 'name' is considered as Knative service
		{"myksvc", &duckv1.KReference{Kind: "Service",
			APIVersion: "serving.knative.dev/v1",
			Namespace:  "default",
			Name:       "myksvc"}, ""},
		// Test 'type:name' format
		{"ksvc:myksvc", &duckv1.KReference{Kind: "Service",
			APIVersion: "serving.knative.dev/v1",
			Namespace:  "default",
			Name:       "myksvc"}, ""},
		{"kroute:mykroute", &duckv1.KReference{Kind: "Route",
			APIVersion: "serving.knative.dev/v1",
			Namespace:  "default",
			Name:       "mykroute"}, ""},
		// Test 'type:name:namespace' format
		{"ksvc:myksvc:other", &duckv1.KReference{Kind: "Service",
			APIVersion: "serving.knative.dev/v1",
			Namespace:  "other",
			Name:       "myksvc"}, ""},
		{"kroute:mykroute:other", &duckv1.KReference{Kind: "Route",
			APIVersion: "serving.knative.dev/v1",
			Namespace:  "other",
			Name:       "mykroute"}, ""},

		{"k8ssvc:foo", nil, "unsupported sink prefix: 'k8ssvc'"},
		{"svc:foo", nil, "unsupported sink prefix: 'svc'"},
		{"service:foo", nil, "unsupported sink prefix: 'service'"},
	}
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("default", myksvc, mykroute, myksvcInOther, mykrouteInOther)
	for _, c := range cases {
		i := &RefFlags{reference: c.ref}
		result, err := i.Resolve(context.Background(), dynamicClient, "default")
		if c.destination != nil {
			assert.DeepEqual(t, result, c.destination)
			assert.NilError(t, err)
		} else {
			assert.ErrorContains(t, err, c.errContents)
		}
	}
}

func TestRefFlagAdd(t *testing.T) {
	c := &cobra.Command{Use: "reftest"}
	refFlag := new(RefFlags)
	refFlag.Add(c)
	assert.Equal(t, "ref", c.Flag("ref").Name)
}

func executeDomainCommand(client clientservingv1beta1.KnServingClient, dynamicClient kndynamic.KnDynamicClient, args ...string) (string, error) {
	knParams := &commands.KnParams{}
	knParams.ClientConfig = blankConfig

	output := new(bytes.Buffer)
	knParams.Output = output
	knParams.NewServingV1beta1Client = func(namespace string) (clientservingv1beta1.KnServingClient, error) {
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

func createDomainMapping(name string, ref duckv1.KReference, tls string) *servingv1beta1.DomainMapping {
	return clientservingv1beta1.NewDomainMappingBuilder(name).Namespace("default").Reference(ref).TLS(tls).Build()
}

func createServiceRef(service, namespace string) duckv1.KReference {
	return duckv1.KReference{Name: service,
		Kind:       "Service",
		APIVersion: "serving.knative.dev/v1",
		Namespace:  namespace,
	}
}
