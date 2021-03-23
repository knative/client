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
	clientdynamic "knative.dev/client/pkg/dynamic"
	"knative.dev/client/pkg/kn/commands"
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
	"svc": {
		Resource: "services",
		Group:    "",
		Version:  "v1",
	},
}

func (f *RefFlags) Add(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.reference, "ref", "", "")
	cmd.Flag("ref").Usage = "Addressable target reference for Domain Mapping. " +
		"You can specify a Knative service, a Knative soute or a Kubernetes service." +
		"Examples: '--ref ksvc:hello' or simply '--ref hello' for a Knative service 'hello', " +
		"'--ref kroute:hello' for a Knative route 'hello', " +
		"'--ref svc:hello' for a Kubernetes service 'hello'. " +
		"If a prefix is not provided, it is considered as a Knative service."
}

func (f RefFlags) Resolve(knclient clientdynamic.KnDynamicClient, namespace string) (*duckv1.KReference, error) {
	client := knclient.RawClient()
	if f.reference == "" {
		return nil, nil
	}

	prefix, name := parseType(f.reference)
	gvr, ok := refMappings[prefix]
	if !ok {
		return nil, fmt.Errorf("unsupported sink prefix: '%s'", prefix)
	}
	obj, err := client.Resource(gvr).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
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

func parseType(ref string) (string, string) {
	parts := strings.SplitN(ref, ":", 2)
	if len(parts) == 1 {
		return "ksvc", parts[0]
	} else {
		return parts[0], parts[1]
	}
}
