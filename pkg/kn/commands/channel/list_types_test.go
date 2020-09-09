/*
Copyright © 2020 The Knative Authors

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
package channel

import (
	"strings"
	"testing"

	"gotest.tools/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"

	dynamicfake "k8s.io/client-go/dynamic/fake"
	"knative.dev/client/pkg/dynamic"
	"knative.dev/client/pkg/kn/commands"

	"knative.dev/client/pkg/util"
)

const (
	crdGroup           = "apiextensions.k8s.io"
	crdVersion         = "v1beta1"
	crdKind            = "CustomResourceDefinition"
	crdKinds           = "customresourcedefinitions"
	testNamespace      = "current"
	channelLabelKey    = "messaging.knative.dev/subscribable"
	channelLabelValue  = "true"
	channelListGroup   = "messaging.knative.dev"
	channelListVersion = "v1beta1"
	channelListKind    = "ChannelList"
)

// channelFakeCmd takes cmd to be executed using dynamic client
// pass the objects to be registered to dynamic client
func channelFakeCmd(args []string, objects ...runtime.Object) (output []string, err error) {
	knParams := &commands.KnParams{}
	cmd, _, buf := commands.CreateDynamicTestKnCommand(NewChannelCommand(knParams), knParams, objects...)
	cmd.SetArgs(args)
	err = cmd.Execute()
	if err != nil {
		return
	}
	output = strings.Split(buf.String(), "\n")
	return
}

func TestChannelListTypesNoChannelInstalled(t *testing.T) {
	_, err := channelFakeCmd([]string{"channel", "list-types"})
	assert.Check(t, err != nil)
	assert.Check(t, util.ContainsAll(err.Error(), "No channel found on the backend, please verify the installation"))
}

func TestChannelListTypes(t *testing.T) {
	output, err := channelFakeCmd([]string{"channel", "list-types"},
		newChannelCRDObjWithSpec("InMemoryChannel", "messaging.knative.dev", "v1beta1", "InMemoryChannel"),
	)
	assert.NilError(t, err)
	assert.Check(t, util.ContainsAll(output[0], "TYPE", "GROUP", "VERSION"))
	assert.Check(t, util.ContainsAll(output[1], "InMemoryChannel", "messaging.knative.dev", "v1beta1"))
}

func TestChannelListTypesNoHeaders(t *testing.T) {
	output, err := channelFakeCmd([]string{"channel", "list-types", "--no-headers"},
		newChannelCRDObjWithSpec("InMemoryChannel", "messaging.knative.dev", "v1beta1", "InMemoryChannel"),
	)
	assert.NilError(t, err)
	assert.Check(t, util.ContainsNone(output[0], "TYPE", "GROUP", "VERSION"))
	assert.Check(t, util.ContainsAll(output[0], "InMemoryChannel", "messaging.knative.dev", "v1beta1"))
}

func TestListBuiltInChannelTypes(t *testing.T) {
	fakeDynamic := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())
	channel, err := listBuiltInChannelTypes(dynamic.NewKnDynamicClient(fakeDynamic, "current"))
	assert.NilError(t, err)
	assert.Check(t, channel != nil)
	assert.Equal(t, len(channel.Items), 1)
}

func newChannelCRDObjWithSpec(name, group, version, kind string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": crdGroup + "/" + crdVersion,
			"kind":       crdKind,
			"metadata": map[string]interface{}{
				"namespace": testNamespace,
				"name":      name,
			},
		},
	}
	obj.Object["spec"] = map[string]interface{}{
		"group":   group,
		"version": version,
		"names": map[string]interface{}{
			"kind":   kind,
			"plural": strings.ToLower(kind) + "s",
		},
	}
	obj.SetLabels(labels.Set{channelLabelKey: channelLabelValue})
	return obj
}
