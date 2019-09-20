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
			client_s, err := p.NewClient(namespaceS)
			if err != nil {
				return err
			}

			err = printServiceWithRevisions(client_s, namespaceS, "source")
			if err != nil {
				return err
			}

			dp := commands.KnParams{
				KubeCfgPath: kubeconfigD,
			}
			// For destination
			dp.Initialize()
			client_d, err := dp.NewClient(namespaceD)
			if err != nil {
				return err
			}

			fmt.Println("[Before migration in destination cluster]")
			err = printServiceWithRevisions(client_d, namespaceD, "destination")
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

			services_s, err := client_s.ListServices()
			if err != nil {
				return err
			}
			for i := 0; i < len(services_s.Items); i++ {
				service_s := services_s.Items[i]
				serviceExists, err := serviceExists(client_d, service_s.Name)
				if err != nil {
					return err
				}

				if serviceExists {
					if !migrateFlags.ForceReplace {
						fmt.Println("\n[Error] Cannot migrate service", service_s.Name, "in namespace", namespaceS,
							"because the service already exists and no --force option was given")
						os.Exit(1)
					}
					fmt.Println("Deleting service", service_s.Name, "from the destination cluster and recreate as replacement")
					err = client_d.DeleteService(service_s.Name)
					if err != nil {
						return err
					}
				}

				err = client_d.CreateService(constructMigratedService(service_s, namespaceD))
				if err != nil {
					return err
				}
				fmt.Println("Migrated service", service_s.Name, "Successfully")

				service_d, err := client_d.GetService(service_s.Name)
				if err != nil {
					return err
				}

				config, err := client_d.GetConfiguration(service_d.Name)
				if err != nil {
					return err
				}
				config_uuid := config.UID

				revisions_s, err := client_s.ListRevisions(v1alpha12.WithService(service_s.Name))
				if err != nil {
					fmt.Errorf(err.Error())
				}
				servingclient_d, err := dp.GetConfig()
				if err != nil {
					return err
				}
				for i := 0; i < len(revisions_s.Items); i++ {
					revision_s := revisions_s.Items[i]
					if revision_s.Name != service_s.Status.LatestReadyRevisionName {
						revision, err := servingclient_d.Revisions(namespaceD).Create(constructRevision(revision_s, config_uuid, namespaceD))
						if err != nil {
							return err
						}
						fmt.Println("Migrated revision", revision.Name, "successfully")
					} else {
						retries := 0
						for {
							revision, err := client_d.GetRevision(revision_s.Name)
							if err != nil {
								return err
							}

							source_revision_generation := revision_s.ObjectMeta.Labels["serving.knative.dev/configurationGeneration"]
							revision.ObjectMeta.Labels["serving.knative.dev/configurationGeneration"] = source_revision_generation
							_, err = servingclient_d.Revisions(namespaceD).Update(revision)
							if err != nil {
								// Retry to update when a resource version conflict exists
								if api_errors.IsConflict(err) && retries < MaxUpdateRetries {
									retries++
									continue
								}
								return err
							}
							fmt.Println("Replace revision", revision_s.Name, "to generation", source_revision_generation, "successfully")
							break
						}
					}
				}
				fmt.Println("")
			}

			fmt.Println("[After migration in destination cluster]")
			err = printServiceWithRevisions(client_d, namespaceD, "destination")
			if err != nil {
				return err
			}

			if cmd.Flag("delete").Value.String() == "false" {
				fmt.Println("Migrate without --delete option, skip deleting Knative resource in source cluster")
			} else {
				fmt.Println("Migrate with --delete option, deleting all Knative resource in source cluster")
				services_s, err := client_s.ListServices()
				if err != nil {
					return err
				}
				for i := 0; i < len(services_s.Items); i++ {
					service_s := services_s.Items[i]
					err = client_s.DeleteService(service_s.Name)
					if err != nil {
						return err
					}
					fmt.Println("Deleted service", service_s.Name, "in source cluster", namespaceS, "namespace")
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
func printServiceWithRevisions(client v1alpha1.KnClient, namespace, clustername string) error {
	services, err := client.ListServices()
	if err != nil {
		return err
	}

	fmt.Println("There are", len(services.Items), "service(s) in", clustername, "cluster", namespace, "namespace:")
	for i := 0; i < len(services.Items); i++ {
		service := services.Items[i]
		fmt.Printf("%-25s%-30s%-20s\n", "Name", "Current Revision", "Ready")
		fmt.Printf("%-25s%-30s%-20s\n", service.Name, service.Status.LatestReadyRevisionName, fmt.Sprint(service.Status.IsReady()))

		revisions_s, err := client.ListRevisions(v1alpha12.WithService(service.Name))
		if err != nil {
			return err
		}
		for i := 0; i < len(revisions_s.Items); i++ {
			revision_s := revisions_s.Items[i]
			fmt.Println("  |- Revision", revision_s.Name, "( Generation: "+fmt.Sprint(revision_s.Labels["serving.knative.dev/configurationGeneration"]), ", Ready:", revision_s.Status.IsReady(), ")")
		}
		fmt.Println("")
	}
	return nil
}

// Create service struct from provided options
func constructMigratedService(originalservice serving_v1alpha1_api.Service, namespace string) *serving_v1alpha1_api.Service {
	service := serving_v1alpha1_api.Service{
		ObjectMeta: originalservice.ObjectMeta,
	}

	service.ObjectMeta.Namespace = namespace
	service.Spec = originalservice.Spec
	service.Spec.Template.ObjectMeta.Name = originalservice.Status.LatestCreatedRevisionName
	service.ObjectMeta.ResourceVersion = ""
	return &service
}

// Create revision struct from provided options
func constructRevision(originalrevision serving_v1alpha1_api.Revision, config_uuid types.UID, namespace string) *serving_v1alpha1_api.Revision {

	revision := serving_v1alpha1_api.Revision{
		ObjectMeta: originalrevision.ObjectMeta,
	}

	revision.ObjectMeta.Namespace = namespace
	revision.ObjectMeta.ResourceVersion = ""
	revision.ObjectMeta.OwnerReferences[0].UID = config_uuid
	revision.ObjectMeta.Labels["serving.knative.dev/configurationGeneration"] = originalrevision.ObjectMeta.Labels["serving.knative.dev/configurationGeneration"]
	revision.Spec = originalrevision.Spec

	return &revision
}
