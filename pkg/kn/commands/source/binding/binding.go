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

package binding

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/clientcmd"
	v1alpha2 "knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1alpha2"
	"knative.dev/pkg/tracker"

	"knative.dev/client/pkg/kn/commands"
	clientsourcesv1alpha1 "knative.dev/client/pkg/sources/v1alpha2"
	"knative.dev/client/pkg/util"
)

// NewBindingCommand is the root command for all binding related commands
func NewBindingCommand(p *commands.KnParams) *cobra.Command {
	bindingCmd := &cobra.Command{
		Use:   "binding",
		Short: "Sink binding command group",
	}
	bindingCmd.AddCommand(NewBindingCreateCommand(p))
	bindingCmd.AddCommand(NewBindingUpdateCommand(p))
	bindingCmd.AddCommand(NewBindingDeleteCommand(p))
	bindingCmd.AddCommand(NewBindingListCommand(p))
	bindingCmd.AddCommand(NewBindingDescribeCommand(p))
	return bindingCmd
}

// This var can be used to inject a factory for fake clients when doing
// tests.
var sinkBindingClientFactory func(config clientcmd.ClientConfig, namespace string) (clientsourcesv1alpha1.KnSinkBindingClient, error)

func newSinkBindingClient(p *commands.KnParams, cmd *cobra.Command) (clientsourcesv1alpha1.KnSinkBindingClient, error) {
	namespace, err := p.GetNamespace(cmd)
	if err != nil {
		return nil, err
	}

	if sinkBindingClientFactory != nil {
		config, err := p.GetClientConfig()
		if err != nil {
			return nil, err
		}
		return sinkBindingClientFactory(config, namespace)
	}

	clientConfig, err := p.RestConfig()
	if err != nil {
		return nil, err
	}

	client, err := v1alpha2.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	return clientsourcesv1alpha1.NewKnSourcesClient(client, namespace).SinkBindingClient(), nil
}

func toReference(subjectArg string, namespace string) (*tracker.Reference, error) {
	parts := strings.SplitN(subjectArg, ":", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid subject argument '%s': not in format kind:api/version:nameOrSelector", subjectArg)
	}
	kind := parts[0]
	gv, err := schema.ParseGroupVersion(parts[1])
	if err != nil {
		return nil, err
	}
	reference := &tracker.Reference{
		APIVersion: gv.String(),
		Kind:       kind,
		Namespace:  namespace,
	}
	if !strings.Contains(parts[2], "=") {
		reference.Name = parts[2]
	} else {
		selector, err := parseSelector(parts[2])
		if err != nil {
			return nil, err
		}
		reference.Selector = &metav1.LabelSelector{MatchLabels: selector}
	}
	return reference, nil
}

func parseSelector(labelSelector string) (map[string]string, error) {
	selector := map[string]string{}
	for _, p := range strings.Split(labelSelector, ",") {
		keyValue := strings.SplitN(p, "=", 2)
		if len(keyValue) != 2 {
			return nil, fmt.Errorf("invalid subject label selector '%s', expected format: key1=value,key2=value", labelSelector)
		}
		selector[keyValue[0]] = keyValue[1]
	}
	return selector, nil
}

// subjectToString converts a reference to a string representation
func subjectToString(ref tracker.Reference) string {

	ret := ref.Kind + ":" + ref.APIVersion
	if ref.Name != "" {
		return ret + ":" + ref.Name
	}
	var keyValues []string
	selector := ref.Selector
	if selector != nil {
		for k, v := range selector.MatchLabels {
			keyValues = append(keyValues, k+"="+v)
		}
		return ret + ":" + strings.Join(keyValues, ",")
	}
	return ret
}

// updateCeOverrides updates the values of the --ce-override flags if given
func updateCeOverrides(bindingFlags bindingUpdateFlags, bindingBuilder *clientsourcesv1alpha1.SinkBindingBuilder) error {
	if bindingFlags.ceOverrides != nil {
		ceOverrideMap, err := util.MapFromArray(bindingFlags.ceOverrides, "=")
		if err != nil {
			return err
		}
		bindingBuilder.AddCloudEventOverrides(ceOverrideMap)
	}
	return nil
}
