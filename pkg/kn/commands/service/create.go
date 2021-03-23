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
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"knative.dev/client/pkg/kn/commands"
	servinglib "knative.dev/client/pkg/serving"

	"knative.dev/serving/pkg/apis/serving"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"

	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	clientservingv1 "knative.dev/client/pkg/serving/v1"
)

var create_example = `
  # Create a service 's0' using image knativesamples/helloworld
  kn service create s0 --image knativesamples/helloworld

  # Create a service with multiple environment variables
  kn service create s1 --env TARGET=v1 --env FROM=examples --image knativesamples/helloworld

  # Create or replace a service using --force flag
  # if service 's1' doesn't exist, it's a normal create operation
  kn service create --force s1 --image knativesamples/helloworld

  # Create or replace environment variables of service 's1' using --force flag
  kn service create --force s1 --env TARGET=force --env FROM=examples --image knativesamples/helloworld

  # Create a service with port 8080
  kn service create s2 --port 8080 --image knativesamples/helloworld

  # Create a service with port 8080 and port name h2c
  kn service create s2 --port h2c:8080 --image knativesamples/helloworld

  # Create or replace default resources of a service 's1' using --force flag
  # (earlier configured resource requests and limits will be replaced with default)
  # (earlier configured environment variables will be cleared too if any)
  kn service create --force s1 --image knativesamples/helloworld

  # Create a service with annotation
  kn service create s3 --image knativesamples/helloworld --annotation sidecar.istio.io/inject=false

  # Create a private service (that is a service with no external endpoint)
  kn service create s1 --image knativesamples/helloworld --cluster-local

  # Create a service with 250MB memory, 200m CPU requests and a GPU resource limit
  # [https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/]
  # [https://kubernetes.io/docs/tasks/manage-gpus/scheduling-gpus/]
  kn service create s4gpu --image knativesamples/hellocuda-go --request memory=250Mi,cpu=200m --limit nvidia.com/gpu=1

  # Create the service in offline mode instead of kubernetes cluster
  kn service create gitopstest -n test-ns --image knativesamples/helloworld --target=/user/knfiles
  kn service create gitopstest --image knativesamples/helloworld --target=/user/knfiles/test.yaml
  kn service create gitopstest --image knativesamples/helloworld --target=/user/knfiles/test.json`

func NewServiceCreateCommand(p *commands.KnParams) *cobra.Command {
	var editFlags ConfigurationEditFlags
	var waitFlags commands.WaitFlags

	serviceCreateCommand := &cobra.Command{
		Use:     "create NAME --image IMAGE",
		Short:   "Create a service",
		Example: create_example,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 && editFlags.Filename == "" {
				return errors.New("'service create' requires the service name given as single argument")
			}
			name := ""
			if len(args) == 1 {
				name = args[0]
			}
			if editFlags.PodSpecFlags.Image == "" && editFlags.Filename == "" {
				return errors.New("'service create' requires the image name to run provided with the --image option")
			}

			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			var service *servingv1.Service
			if editFlags.Filename == "" {
				service, err = constructService(cmd, editFlags, name, namespace)
			} else {
				service, err = constructServiceFromFile(cmd, editFlags, name, namespace)
			}
			if err != nil {
				return err
			}
			targetFlag := cmd.Flag("target").Value.String()
			client, err := newServingClient(p, namespace, targetFlag)
			if err != nil {
				return err
			}
			serviceExists, err := serviceExists(cmd.Context(), client, service.Name)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			if serviceExists {
				if !editFlags.ForceCreate {
					return fmt.Errorf(
						"cannot create service '%s' in namespace '%s' "+
							"because the service already exists and no --force option was given", service.Name, namespace)
				}
				err = replaceService(cmd.Context(), client, service, waitFlags, out, targetFlag)
			} else {
				err = createService(cmd.Context(), client, service, waitFlags, out, targetFlag)
			}
			if err != nil {
				return err
			}
			return nil
		},
	}
	commands.AddNamespaceFlags(serviceCreateCommand.Flags(), false)
	commands.AddGitOpsFlags(serviceCreateCommand.Flags())
	editFlags.AddCreateFlags(serviceCreateCommand)
	waitFlags.AddConditionWaitFlags(serviceCreateCommand, commands.WaitDefaultTimeout, "create", "service", "ready")
	return serviceCreateCommand
}

func createService(ctx context.Context, client clientservingv1.KnServingClient, service *servingv1.Service, waitFlags commands.WaitFlags, out io.Writer, targetFlag string) error {
	err := client.CreateService(ctx, service)
	if err != nil {
		return err
	}

	return waitIfRequested(ctx, client, waitFlags, service.Name, "Creating", "created", targetFlag, out)
}

func replaceService(ctx context.Context, client clientservingv1.KnServingClient, service *servingv1.Service, waitFlags commands.WaitFlags, out io.Writer, targetFlag string) error {
	changed, err := prepareAndUpdateService(ctx, client, service)
	if err != nil {
		return err
	}
	if !changed {
		fmt.Fprintf(out, "Service '%s' replaced in namespace '%s' (unchanged).\n", service.Name, client.Namespace(ctx))
		return nil
	}
	return waitIfRequested(ctx, client, waitFlags, service.Name, "Replacing", "replaced", targetFlag, out)
}

func waitIfRequested(ctx context.Context, client clientservingv1.KnServingClient, waitFlags commands.WaitFlags, serviceName string, verbDoing string, verbDone string, targetFlag string, out io.Writer) error {
	if !waitFlags.Wait || targetFlag != "" {
		fmt.Fprintf(out, "Service '%s' %s in namespace '%s'.\n", serviceName, verbDone, client.Namespace(ctx))
		return nil
	}
	fmt.Fprintf(out, "%s service '%s' in namespace '%s':\n", verbDoing, serviceName, client.Namespace(ctx))
	return waitForServiceToGetReady(ctx, client, serviceName, waitFlags.TimeoutInSeconds, verbDone, out)
}

func prepareAndUpdateService(ctx context.Context, client clientservingv1.KnServingClient, service *servingv1.Service) (bool, error) {
	var retries = 0
	for {
		existingService, err := client.GetService(ctx, service.Name)
		if err != nil {
			return false, err
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
		changed, err := client.UpdateService(ctx, service)
		if err != nil {
			// Retry to update when a resource version conflict exists
			if apierrors.IsConflict(err) && retries < MaxUpdateRetries {
				retries++
				continue
			}
			return changed, err
		}
		return changed, nil
	}
}

func waitForServiceToGetReady(ctx context.Context, client clientservingv1.KnServingClient, name string, timeout int, verbDone string, out io.Writer) error {
	fmt.Fprintln(out, "")
	err := waitForService(ctx, client, name, out, timeout)
	if err != nil {
		return err
	}
	fmt.Fprintln(out, "")
	return showUrl(ctx, client, name, "", verbDone, out)
}

func serviceExists(ctx context.Context, client clientservingv1.KnServingClient, name string) (bool, error) {
	_, err := client.GetService(ctx, name)
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

	if name == "" || namespace == "" {
		return nil, errors.New("internal: no name or namespace provided when constructing a service")
	}

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

// constructServiceFromFile creates struct from provided file
func constructServiceFromFile(cmd *cobra.Command, editFlags ConfigurationEditFlags, name, namespace string) (*servingv1.Service, error) {
	var service servingv1.Service
	file, err := os.Open(editFlags.Filename)
	if err != nil {
		return nil, err
	}
	decoder := yaml.NewYAMLOrJSONDecoder(file, 512)

	err = decoder.Decode(&service)
	if err != nil {
		return nil, err
	}
	if name == "" && service.Name != "" {
		// keep provided service.Name if name param is empty
	} else if name != "" && service.Name == "" {
		service.Name = name
	} else if name != "" && service.Name != "" {
		// throw error if names differ, otherwise use already set value
		if name != service.Name {
			return nil, fmt.Errorf("provided service name '%s' doesn't match name from file '%s'", name, service.Name)
		}
	} else {
		return nil, fmt.Errorf("no service name provided in command parameter or file")
	}

	// Set namespace in case it's specified as --namespace
	service.ObjectMeta.Namespace = namespace

	// Apply options provided from cmdline
	err = editFlags.Apply(&service, nil, cmd)
	if err != nil {
		return nil, err
	}

	return &service, nil
}
