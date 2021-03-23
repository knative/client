// Copyright © 2020 The Knative Authors
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

package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/util/retry"

	clientv1alpha1 "knative.dev/client/pkg/apis/client/v1alpha1"
	"knative.dev/client/pkg/kn/commands"
	clientservingv1 "knative.dev/client/pkg/serving/v1"
	"knative.dev/pkg/kmeta"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

// NewServiceImportCommand returns a new command for importing a service.
func NewServiceImportCommand(p *commands.KnParams) *cobra.Command {
	var waitFlags commands.WaitFlags

	command := &cobra.Command{
		Use:   "import FILENAME",
		Short: "Import a service and its revisions (experimental)",
		Example: `
 # Import a service from YAML file
 kn service import /path/to/file.yaml

 # Import a service from JSON file
 kn service import /path/to/file.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("'kn service import' requires filename of import file as single argument")
			}
			filename := args[0]

			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			client, err := p.NewServingClient(namespace)
			if err != nil {
				return err
			}

			return importWithOwnerRef(cmd.Context(), client, filename, cmd.OutOrStdout(), waitFlags)
		},
	}
	flags := command.Flags()
	commands.AddNamespaceFlags(flags, false)
	waitFlags.AddConditionWaitFlags(command, commands.WaitDefaultTimeout, "import", "service", "ready")

	return command
}

func importWithOwnerRef(ctx context.Context, client clientservingv1.KnServingClient, filename string, out io.Writer, waitFlags commands.WaitFlags) error {
	var export clientv1alpha1.Export
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	decoder := yaml.NewYAMLOrJSONDecoder(file, 512)
	err = decoder.Decode(&export)
	if err != nil {
		return err
	}
	if export.Spec.Service.Name == "" {
		return fmt.Errorf("provided import file doesn't contain service name, please note that only kn's custom export format is supported")
	}

	serviceName := export.Spec.Service.Name

	// Return error if service already exists
	svcExists, err := serviceExists(ctx, client, serviceName)
	if err != nil {
		return err
	}
	if svcExists {
		return fmt.Errorf("cannot import service '%s' in namespace '%s' because the service already exists",
			serviceName, client.Namespace(ctx))
	}

	err = client.CreateService(ctx, &export.Spec.Service)
	if err != nil {
		return err
	}

	// Retrieve current Configuration to be use in OwnerReference
	currentConf, err := getConfigurationWithRetry(ctx, client, serviceName)
	if err != nil {
		return err
	}

	// Create revision with current Configuration's OwnerReference
	if len(export.Spec.Revisions) > 0 {
		for _, r := range export.Spec.Revisions {
			tmp := r.DeepCopy()
			// OwnerRef ensures that Revisions are recognized by controller
			tmp.OwnerReferences = []metav1.OwnerReference{*kmeta.NewControllerRef(currentConf)}
			if err = client.CreateRevision(ctx, tmp); err != nil {
				return err
			}
		}
	}

	err = waitIfRequested(ctx, client, waitFlags, serviceName, "Importing", "imported", "", out)
	if err != nil {
		return err
	}
	return err
}

func getConfigurationWithRetry(ctx context.Context, client clientservingv1.KnServingClient, name string) (*servingv1.Configuration, error) {
	var conf *servingv1.Configuration
	var err error
	err = retry.OnError(retry.DefaultBackoff, func(err error) bool {
		return apierrors.IsNotFound(err)
	}, func() error {
		conf, err = client.GetConfiguration(ctx, name)
		return err
	})
	return conf, err
}
