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
	"fmt"
	"os"

	"github.com/spf13/cobra"
	apiv1 "k8s.io/api/core/v1"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clientset "k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // from https://github.com/kubernetes/client-go/issues/345
	"k8s.io/client-go/tools/clientcmd"
	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/serving/v1alpha1"
	v1alpha12 "knative.dev/client/pkg/serving/v1alpha1"
	serving_v1alpha1_api "knative.dev/serving/pkg/apis/serving/v1alpha1"
)

func NewServiceMigrateCommand(p *commands.KnParams) *cobra.Command {
	var migrateFlags MigrateFlags

	serviceMigrateCommand := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate Knative services from source cluster to destination cluster",
		Example: `
  # Migrate Knative services from source cluster to destination cluster by export KUBECONFIG and KUBECONFIG_DESTINATION as environment variables
  kn migrate --namespace default --destination-namespace default

  # Migrate Knative services from source cluster to destination cluster by set kubeconfig as parameters
  kn migrate --namespace default --destination-namespace default --kubeconfig $HOME/.kube/config/source-cluster-config.yml --destination-kubeconfig $HOME/.kube/config/destination-cluster-config.yml

  # Migrate Knative services from source cluster to destination cluster and force replace the service if exists in destination cluster
  kn migrate --namespace default --destination-namespace default --force

  # Migrate Knative services from source cluster to destination cluster and delete the service in source cluster
  kn migrate --namespace default --destination-namespace default --force --delete`,

		RunE: func(cmd *cobra.Command, args []string) (err error) {

			namespaceS := ""
			namespaceD := ""
			if migrateFlags.SourceNamespace == "" {
				return fmt.Errorf("cannot get source cluster namespace, please use --namespace to set")
			} else {
				namespaceS = migrateFlags.SourceNamespace
			}
			p.GetClientConfig()

			if migrateFlags.DestinationNamespace == "" {
				return fmt.Errorf("cannot get destination cluster namespace, please use --destination-namespace to set")
			} else {
				namespaceD = migrateFlags.DestinationNamespace
			}

			kubeconfigS := p.KubeCfgPath
			if kubeconfigS == "" {
				kubeconfigS = os.Getenv("KUBECONFIG")
			}
			if kubeconfigS == "" {
				return fmt.Errorf("cannot get source cluster kube config, please use --kubeconfig or export environment variable KUBECONFIG to set")
			}

			kubeconfigD := migrateFlags.DestinationKubeconfig
			if kubeconfigD == "" {
				kubeconfigD = os.Getenv("KUBECONFIG_DESTINATION")
			}
			if kubeconfigD == "" {
				return fmt.Errorf("cannot get destination cluster kube config, please use --destination-kubeconfig or export environment variable KUBECONFIG_DESTINATION to set")
			}

			// For source
			p.KubeCfgPath = kubeconfigS
			clientS, err := p.NewServingClient(namespaceS)
			if err != nil {
				return err
			}

			err = printServiceWithRevisions(clientS, namespaceS, "source")
			if err != nil {
				return err
			}

			dp := commands.KnParams{
				KubeCfgPath: kubeconfigD,
			}
			// For destination
			dp.Initialize()
			clientD, err := dp.NewServingClient(namespaceD)
			if err != nil {
				return err
			}

			fmt.Println("[Before migration in destination cluster]")
			err = printServiceWithRevisions(clientD, namespaceD, "destination")
			if err != nil {
				return err
			}

			fmt.Println("\nNow migrate all Knative resources: \nFrom the source namespace ", namespaceS, "of cluster", p.KubeCfgPath)
			fmt.Println("To the destination namespace", namespaceD, "of cluster", kubeconfigD)

			cfg_d, err := clientcmd.BuildConfigFromFlags("", dp.KubeCfgPath)
			clientset, err := clientset.NewForConfig(cfg_d)
			if err != nil {
				return err
			}
			namespaceExists, err := namespaceExists(*clientset, namespaceD)
			if err != nil {
				return err
			}

			if !namespaceExists {
				fmt.Println("Create namespace", namespaceD, "in destination cluster")
				nsSpec := &apiv1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespaceD}}
				_, err = clientset.CoreV1().Namespaces().Create(nsSpec)
				if err != nil {
					return err
				}
			} else {
				fmt.Println("Namespace", namespaceD, "already exists in destination cluster")
			}
			if err != nil {
				return err
			}

			servicesS, err := clientS.ListServices()
			if err != nil {
				return err
			}
			for i := 0; i < len(servicesS.Items); i++ {
				serviceS := servicesS.Items[i]
				serviceExists, err := serviceExists(clientD, serviceS.Name)
				if err != nil {
					return err
				}

				if serviceExists {
					if !migrateFlags.ForceReplace {
						fmt.Println("\n[Error] Cannot migrate service", serviceS.Name, "in namespace", namespaceS,
							"because the service already exists and no --force option was given")
						os.Exit(1)
					}
					fmt.Println("Deleting service", serviceS.Name, "from the destination cluster and recreate as replacement")
					err = clientD.DeleteService(serviceS.Name)
					if err != nil {
						return err
					}
				}

				err = clientD.CreateService(constructMigratedService(serviceS, namespaceD))
				if err != nil {
					return err
				}
				fmt.Println("Migrated service", serviceS.Name, "Successfully")

				serviceD, err := clientD.GetService(serviceS.Name)
				if err != nil {
					return err
				}

				config, err := clientD.GetConfiguration(serviceD.Name)
				if err != nil {
					return err
				}
				configUuid := config.UID

				revisionsS, err := clientS.ListRevisions(v1alpha12.WithService(serviceS.Name))
				if err != nil {
					fmt.Errorf(err.Error())
				}

				if err != nil {
					return err
				}
				for i := 0; i < len(revisionsS.Items); i++ {
					revisionS := revisionsS.Items[i]
					if revisionS.Name != serviceS.Status.LatestReadyRevisionName {
						err := clientD.CreateRevision(constructRevision(revisionS, configUuid, namespaceD))
						if err != nil {
							return err
						}
						fmt.Println("Migrated revision", revisionS.Name, "successfully")
					} else {
						retries := 0
						for {
							revision, err := clientD.GetRevision(revisionS.Name)
							if err != nil {
								return err
							}

							sourceRevisionGeneration := revisionS.ObjectMeta.Labels["serving.knative.dev/configurationGeneration"]
							revision.ObjectMeta.Labels["serving.knative.dev/configurationGeneration"] = sourceRevisionGeneration
							err = clientD.UpdateRevision(revision)
							if err != nil {
								// Retry to update when a resource version conflict exists
								if api_errors.IsConflict(err) && retries < MaxUpdateRetries {
									retries++
									continue
								}
								return err
							}
							fmt.Println("Replace revision", revisionS.Name, "to generation", sourceRevisionGeneration, "successfully")
							break
						}
					}
				}
				fmt.Println("")
			}

			fmt.Println("[After migration in destination cluster]")
			err = printServiceWithRevisions(clientD, namespaceD, "destination")
			if err != nil {
				return err
			}

			if cmd.Flag("delete").Value.String() == "false" {
				fmt.Println("Migrate without --delete option, skip deleting Knative resource in source cluster")
			} else {
				fmt.Println("Migrate with --delete option, deleting all Knative resource in source cluster")
				servicesS, err := clientS.ListServices()
				if err != nil {
					return err
				}
				for i := 0; i < len(servicesS.Items); i++ {
					serviceS := servicesS.Items[i]
					err = clientS.DeleteService(serviceS.Name)
					if err != nil {
						return err
					}
					fmt.Println("Deleted service", serviceS.Name, "in source cluster", namespaceS, "namespace")
				}
			}
			return nil
		},
	}
	migrateFlags.addFlags(serviceMigrateCommand)
	return serviceMigrateCommand
}

func namespaceExists(client clientset.Clientset, namespace string) (bool, error) {
	_, err := client.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{})
	if api_errors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Get service list with revisions
func printServiceWithRevisions(client v1alpha1.KnServingClient, namespace, clustername string) error {
	services, err := client.ListServices()
	if err != nil {
		return err
	}

	fmt.Println("There are", len(services.Items), "service(s) in", clustername, "cluster", namespace, "namespace:")
	for i := 0; i < len(services.Items); i++ {
		service := services.Items[i]
		fmt.Printf("%-25s%-30s%-20s\n", "Name", "Current Revision", "Ready")
		fmt.Printf("%-25s%-30s%-20s\n", service.Name, service.Status.LatestReadyRevisionName, fmt.Sprint(service.Status.IsReady()))

		revisionsS, err := client.ListRevisions(v1alpha12.WithService(service.Name))
		if err != nil {
			return err
		}
		for i := 0; i < len(revisionsS.Items); i++ {
			revisionS := revisionsS.Items[i]
			fmt.Println("  |- Revision", revisionS.Name, "( Generation: "+fmt.Sprint(revisionS.Labels["serving.knative.dev/configurationGeneration"]), ", Ready:", revisionS.Status.IsReady(), ")")
		}
		fmt.Println("")
	}
	return nil
}

// Create service struct from provided options
func constructMigratedService(originalService serving_v1alpha1_api.Service, namespace string) *serving_v1alpha1_api.Service {
	service := serving_v1alpha1_api.Service{
		ObjectMeta: originalService.ObjectMeta,
	}

	service.ObjectMeta.Namespace = namespace
	service.Spec = originalService.Spec
	service.Spec.Template.ObjectMeta.Name = originalService.Status.LatestCreatedRevisionName
	service.ObjectMeta.ResourceVersion = ""
	return &service
}

// Create revision struct from provided options
func constructRevision(originalRevision serving_v1alpha1_api.Revision, configUuid types.UID, namespace string) *serving_v1alpha1_api.Revision {

	revision := serving_v1alpha1_api.Revision{
		ObjectMeta: originalRevision.ObjectMeta,
	}

	revision.ObjectMeta.Namespace = namespace
	revision.ObjectMeta.ResourceVersion = ""
	revision.ObjectMeta.OwnerReferences[0].UID = configUuid
	revision.ObjectMeta.Labels["serving.knative.dev/configurationGeneration"] = originalRevision.ObjectMeta.Labels["serving.knative.dev/configurationGeneration"]
	revision.Spec = originalRevision.Spec

	return &revision
}
