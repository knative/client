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

	"knative.dev/client/pkg/kn/commands"
	servinglib "knative.dev/client/pkg/serving"

	"knative.dev/serving/pkg/apis/serving"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	clientservingv1 "knative.dev/client/pkg/serving/v1"
)

var create_example = `
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
  kn service create --force s1 --image dev.local/ns/image:v1

  # Create a service with annotation
  kn service create s1 --image dev.local/ns/image:v3 --annotation sidecar.istio.io/inject=false

  # Create a private service (that is a service with no external endpoint)
  kn service create s1 --image dev.local/ns/image:v3 --cluster-local`

func NewServiceCreateCommand(p *commands.KnParams) *cobra.Command {
	var editFlags ConfigurationEditFlags
	var waitFlags commands.WaitFlags

	serviceCreateCommand := &cobra.Command{
		Use:     "create NAME --image IMAGE",
		Short:   "Create a service.",
		Example: create_example,
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

			client, err := p.NewServingClient(namespace)
			if err != nil {
				return err
			}

			serviceExists, err := serviceExists(client, name)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			if serviceExists {
				if !editFlags.ForceCreate {
					return fmt.Errorf(
						"cannot create service '%s' in namespace '%s' "+
							"because the service already exists and no --force option was given", name, namespace)
				}
				err = replaceService(client, service, waitFlags, out)
			} else {
				err = createService(client, service, waitFlags, out)
			}
			if err != nil {
				return err
			}
			return nil
		},
	}
	commands.AddNamespaceFlags(serviceCreateCommand.Flags(), false)
	editFlags.AddCreateFlags(serviceCreateCommand)
	waitFlags.AddConditionWaitFlags(serviceCreateCommand, commands.WaitDefaultTimeout, "Create", "service", "ready")
	return serviceCreateCommand
}

func createService(client clientservingv1.KnServingClient, service *servingv1.Service, waitFlags commands.WaitFlags, out io.Writer) error {
	err := client.CreateService(service)
	if err != nil {
		return err
	}

	return waitIfRequested(client, service, waitFlags, "Creating", "created", out)
}

func replaceService(client clientservingv1.KnServingClient, service *servingv1.Service, waitFlags commands.WaitFlags, out io.Writer) error {
	err := prepareAndUpdateService(client, service)
	if err != nil {
		return err
	}
	return waitIfRequested(client, service, waitFlags, "Replacing", "replaced", out)
}

func waitIfRequested(client clientservingv1.KnServingClient, service *servingv1.Service, waitFlags commands.WaitFlags, verbDoing string, verbDone string, out io.Writer) error {
	//TODO: deprecated condition should be removed with --async flag
	if waitFlags.Async {
		fmt.Fprintf(out, "\nWARNING: flag --async is deprecated and going to be removed in future release, please use --no-wait instead.\n\n")
		fmt.Fprintf(out, "Service '%s' %s in namespace '%s'.\n", service.Name, verbDone, client.Namespace())
		return nil
	}
	if waitFlags.NoWait {
		fmt.Fprintf(out, "Service '%s' %s in namespace '%s'.\n", service.Name, verbDone, client.Namespace())
		return nil
	}

	fmt.Fprintf(out, "%s service '%s' in namespace '%s':\n", verbDoing, service.Name, client.Namespace())
	return waitForServiceToGetReady(client, service.Name, waitFlags.TimeoutInSeconds, verbDone, out)
}

func prepareAndUpdateService(client clientservingv1.KnServingClient, service *servingv1.Service) error {
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
			if apierrors.IsConflict(err) && retries < MaxUpdateRetries {
				retries++
				continue
			}
			return err
		}
		return nil
	}
}

func waitForServiceToGetReady(client clientservingv1.KnServingClient, name string, timeout int, verbDone string, out io.Writer) error {
	fmt.Fprintln(out, "")
	err := waitForService(client, name, out, timeout)
	if err != nil {
		return err
	}
	fmt.Fprintln(out, "")
	return showUrl(client, name, "", verbDone, out)
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

func serviceExists(client clientservingv1.KnServingClient, name string) (bool, error) {
	_, err := client.GetService(name)
	if apierrors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Create service struct from provided options
func constructService(cmd *cobra.Command, editFlags ConfigurationEditFlags, name string, namespace string) (*servingv1.Service,
	error) {

	service := servingv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	service.Spec.Template = servingv1.RevisionTemplateSpec{
		Spec: servingv1.RevisionSpec{},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				servinglib.UserImageAnnotationKey: "", // Placeholder. Will be replaced or deleted as we apply mutations.
			},
		},
	}
	service.Spec.Template.Spec.Containers = []corev1.Container{{}}

	err := editFlags.Apply(&service, nil, cmd)
	if err != nil {
		return nil, err
	}
	return &service, nil
}
