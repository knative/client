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
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/client/pkg/commands"
	clientdynamic "knative.dev/client/pkg/dynamic"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

// NewDomainCommand to manage domain mappings
func NewDomainCommand(p *commands.KnParams) *cobra.Command {
	domainCmd := &cobra.Command{
		Use:     "domain COMMAND",
		Short:   "Manage domain mappings",
		Aliases: []string{"domains"},
	}
	domainCmd.AddCommand(NewDomainMappingCreateCommand(p))
	domainCmd.AddCommand(NewDomainMappingDescribeCommand(p))
	domainCmd.AddCommand(NewDomainMappingUpdateCommand(p))
	domainCmd.AddCommand(NewDomainMappingDeleteCommand(p))
	domainCmd.AddCommand(NewDomainMappingListCommand(p))
	return domainCmd
}

type RefFlags struct {
	reference string
	tls       string
}

var refMappings = map[string]schema.GroupVersionResource{
	"ksvc": {
		Resource: "services",
		Group:    "serving.knative.dev",
		Version:  "v1",
	},
	"kroute": {
		Resource: "routes",
		Group:    "serving.knative.dev",
		Version:  "v1",
	},
}

func (f *RefFlags) Add(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.reference, "ref", "", "")
	cmd.Flag("ref").Usage = "Addressable target reference for Domain Mapping. " +
		"You can specify a Knative service, a Knative route. " +
		"Examples: '--ref' ksvc:hello' or simply '--ref hello' for a Knative service 'hello', " +
		"'--ref' kroute:hello' for a Knative route 'hello'. " +
		"'--ref ksvc:mysvc:mynamespace' for a Knative service 'mysvc' in another namespace 'mynamespace', " +
		"If a prefix is not provided, it is considered as a Knative service in the current namespace. " +
		"If referring to a Knative service in another namespace, 'ksvc:name:namespace' combination must be provided explicitly."
}

func (f RefFlags) Resolve(ctx context.Context, knclient clientdynamic.KnDynamicClient, namespace string) (*duckv1.KReference, error) {
	client := knclient.RawClient()
	if f.reference == "" {
		return nil, nil
	}

	prefix, name, refNamespace := parseType(f.reference)
	gvr, ok := refMappings[prefix]
	if !ok {
		return nil, fmt.Errorf("unsupported sink prefix: '%s'", prefix)
	}
	if refNamespace != "" {
		namespace = refNamespace
	}
	obj, err := client.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	result := &duckv1.KReference{
		Kind:       obj.GetKind(),
		APIVersion: obj.GetAPIVersion(),
		Name:       obj.GetName(),
		Namespace:  namespace,
	}
	return result, nil
}

func parseType(ref string) (string, string, string) {
	parts := strings.SplitN(ref, ":", 3)
	switch {
	case len(parts) == 1:
		return "ksvc", parts[0], ""
	case len(parts) == 3:
		return parts[0], parts[1], parts[2]
	default:
		return parts[0], parts[1], ""
	}
}
