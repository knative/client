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
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	v1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/apis/duck/v1beta1"
	"knative.dev/pkg/tracker"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/flags"
	v1alpha12 "knative.dev/client/pkg/sources/v1alpha1"
)

// NewCronJobCreateCommand is for creating CronJob source COs
func NewSinkBindingCommand(p *commands.KnParams) *cobra.Command {
	var bindingFlags bindingUpdateFlags
	var sinkFlags flags.SinkFlags

	cmd := &cobra.Command{
		Use:   "create NAME --subject SCHEDULE --sink SINK --ce-override OVERRIDE",
		Short: "Create a sink binding source.",
		Example: `
  # Create a sink binding source, which connects a deployment 'myapp' with a Knative service 'mysvc''
  kn source binding create my-binding --subject "" --sink svc:mysvc`,

		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("requires the name of the crojob source to create as single argument")

			}
			name := args[0]

			sinkBindingClient, err := newSinkBindingClient(p, cmd)
			if err != nil {
				return err
			}

			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}
			dynamicClient, err := p.NewDynamicClient(namespace)
			if err != nil {
				return err
			}

			destination, err := sinkFlags.ResolveSink(dynamicClient, namespace)
			if err != nil {
				return err
			}

			reference, err := toReference(bindingFlags.subject, bindingFlags.subjectNamespace)
			if err != nil {
				return err
			}

			binding, err := v1alpha12.NewSinkBindingBuilder(name).
				Sink(toDuckV1(destination)).
				Subject(reference).
				Build()
			if err != nil {
				return err
			}
			err = sinkBindingClient.CreateSinkBinding(binding)
			if err == nil {
				fmt.Fprintf(cmd.OutOrStdout(), "Sink binding '%s' created in namespace '%s'.\n", args[0], sinkBindingClient.Namespace())
			}
			return err
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), false)
	bindingFlags.addBindingFlags(cmd)
	sinkFlags.Add(cmd)
	cmd.MarkFlagRequired("subject")
	cmd.MarkFlagRequired("sink")

	return cmd
}

func toReference(subjectArg string, subjectNamespace string) (*tracker.Reference, error) {
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
	}
	if !strings.Contains(parts[2], "=") {
		reference.Name = parts[2]
	} else {
		selector := map[string]string{}
		for _, p := range strings.Split(parts[2], ",") {
			keyValue := strings.SplitN(p, "=", 2)
			if len(keyValue) != 2 {
				return nil, fmt.Errorf("invalid subject label selector '%s' for subject argument %s. format: key1=value,key2=value", parts[2], subjectArg)
			}
			selector[keyValue[0]] = keyValue[1]
		}
		reference.Selector = &metav1.LabelSelector{MatchLabels: selector}
	}
	if subjectNamespace != "" {
		reference.Namespace = subjectNamespace
	}
	return reference, nil
}

// Temporary conversion function until we move to duckv1
func toDuckV1(destination *v1beta1.Destination) *v1.Destination {
	return &v1.Destination{
		Ref: destination.Ref,
		URI: destination.URI,
	}
}
