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

package service

import (
	"errors"
	"fmt"
	"io"

	"github.com/knative/client/pkg/kn/commands"
	"github.com/knative/client/pkg/serving/v1alpha1"
	"github.com/knative/serving/pkg/apis/serving"

	serving_v1alpha1_api "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/spf13/cobra"

	corev1 "k8s.io/api/core/v1"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewServiceCreateCommand(p *commands.KnParams) *cobra.Command {
	var editFlags ConfigurationEditFlags
	var waitFlags commands.WaitFlags

	serviceCreateCommand := &cobra.Command{
		Use:   "create NAME --image IMAGE",
		Short: "Create a service.",
		Example: `
  # Create a service 'mysvc' using image at dev.local/ns/image:latest
  kn service create mysvc --image dev.local/ns/image:latest

  # Create a service with multiple environment variables
  kn service create mysvc --env KEY1=VALUE1 --env KEY2=VALUE2 --image dev.local/ns/image:latest

  # Create or replace a service 's1' with image dev.local/ns/image:v2 using --force flag
  # if service 's1' doesn't exist, it's just a normal create operation
  kn service create --force s1 --image dev.local/ns/image:v2

  # Create or replace environment variables of service 's1' using --force flag
  kn service create --force s1 --env KEY1=NEW_VALUE1 --env NEW_KEY2=NEW_VALUE2 --image dev.local/ns/image:v1

  # Create service 'mysvc' with port 80
  kn service create mysvc --port 80 --image dev.local/ns/image:latest

  # Create or replace default resources of a service 's1' using --force flag
  # (earlier configured resource requests and limits will be replaced with default)
  # (earlier configured environment variables will be cleared too if any)
  kn service create --force s1 --image dev.local/ns/image:v1`,

		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("'service create' requires the service name given as single argument")
			}
			name := args[0]
			if editFlags.Image == "" {
				return errors.New("'service create' requires the image name to run provided with the --image option")
			}

			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			service, err := constructService(cmd, editFlags, name, namespace)
			if err != nil {
				return err
			}

			client, err := p.NewClient(namespace)
			if err != nil {
				return err
			}

			serviceExists, err := serviceExists(client, name, namespace)
			if err != nil {
				return err
			}

			if serviceExists {
				if !editFlags.ForceCreate {
					return fmt.Errorf(
						"cannot create service '%s' in namespace '%s' "+
							"because the service already exists and no --force option was given", name, namespace)
				}
				err = replaceService(client, service, namespace, cmd.OutOrStdout())
			} else {
				err = createService(client, service, namespace, cmd.OutOrStdout())
			}
			if err != nil {
				return err
			}

			if !waitFlags.Async {
				out := cmd.OutOrStdout()
				err := waitForService(client, name, out, waitFlags.TimeoutInSeconds)
				if err != nil {
					return err
				}
				return showUrl(client, name, namespace, out)
			}

			return nil
		},
	}
	commands.AddNamespaceFlags(serviceCreateCommand.Flags(), false)
	editFlags.AddCreateFlags(serviceCreateCommand)
	waitFlags.AddConditionWaitFlags(serviceCreateCommand, 60, "Create", "service")
	return serviceCreateCommand
}

// Duck type for writers having a flush
type flusher interface {
	Flush() error
}

func flush(out io.Writer) {
	if flusher, ok := out.(flusher); ok {
		flusher.Flush()
	}
}

func createService(client v1alpha1.KnClient, service *serving_v1alpha1_api.Service, namespace string, out io.Writer) error {
	err := client.CreateService(service)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Service '%s' successfully created in namespace '%s'.\n", service.Name, namespace)
	return nil
}

func replaceService(client v1alpha1.KnClient, service *serving_v1alpha1_api.Service, namespace string, out io.Writer) error {
	var retries = 0
	for {
		existingService, err := client.GetService(service.Name)
		if err != nil {
			return err
		}

		// Copy over some annotations that we want to keep around. Erase others
		copyList := []string{
			serving.CreatorAnnotation,
			serving.UpdaterAnnotation,
		}

		// If the target Annotation doesn't exist, create it even if
		// we don't end up copying anything over so that we erase all
		// existing annotations
		if service.Annotations == nil {
			service.Annotations = map[string]string{}
		}

		// Do the actual copy now, but only if it's in the source annotation
		for _, k := range copyList {
			if v, ok := existingService.Annotations[k]; ok {
				service.Annotations[k] = v
			}
		}

		service.ResourceVersion = existingService.ResourceVersion
		err = client.UpdateService(service)
		if err != nil {
			// Retry to update when a resource version conflict exists
			if api_errors.IsConflict(err) && retries < MaxUpdateRetries {
				retries++
				continue
			}
			return err
		}
		fmt.Fprintf(out, "Service '%s' successfully replaced in namespace '%s'.\n", service.Name, namespace)
		return nil
	}
}

func serviceExists(client v1alpha1.KnClient, name string, namespace string) (bool, error) {
	_, err := client.GetService(name)
	if api_errors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Create service struct from provided options
func constructService(cmd *cobra.Command, editFlags ConfigurationEditFlags, name string, namespace string) (*serving_v1alpha1_api.Service,
	error) {

	service := serving_v1alpha1_api.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	// TODO: Should it always be `runLatest` ?
	service.Spec.DeprecatedRunLatest = &serving_v1alpha1_api.RunLatestType{
		Configuration: serving_v1alpha1_api.ConfigurationSpec{
			DeprecatedRevisionTemplate: &serving_v1alpha1_api.RevisionTemplateSpec{
				Spec: serving_v1alpha1_api.RevisionSpec{
					DeprecatedContainer: &corev1.Container{},
				},
			},
		},
	}

	err := editFlags.Apply(&service, cmd)
	if err != nil {
		return nil, err
	}
	return &service, nil
}

func showUrl(client v1alpha1.KnClient, serviceName string, namespace string, out io.Writer) error {
	service, err := client.GetService(serviceName)
	if err != nil {
		return fmt.Errorf("cannot fetch service '%s' in namespace '%s' for extracting the URL: %v", serviceName, namespace, err)
	}
	url := service.Status.URL.String()
	if url == "" {
		url = service.Status.DeprecatedDomain
	}
	fmt.Fprintln(out, "\nService URL:")
	fmt.Fprintf(out, "%s\n", url)
	return nil
}
