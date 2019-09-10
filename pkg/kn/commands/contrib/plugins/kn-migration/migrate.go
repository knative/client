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

package kn_migration

import (
	"fmt"
	//"k8s.io/client-go/tools/clientcmd"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // from https://github.com/kubernetes/client-go/issues/345
	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/serving/v1alpha1"
	v1alpha12 "knative.dev/client/pkg/serving/v1alpha1"
	//"k8s.io/client-go/tools/clientcmd"
	//serving_v1alpha1_client "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
)

func NewMigrateCommand(p *commands.KnParams) *cobra.Command {
	var migrateFlags MigrateFlags

	serviceMigrateCommand := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate knative services from source cluster to destination cluster",
		Long:  `Migrate knative services from source cluster to destination cluster`,

		RunE: func(cmd *cobra.Command, args []string) (err error) {

			kubeconfig_d := migrateFlags.DestinationKubeconfig
			if kubeconfig_d == "" {
				kubeconfig_d = os.Getenv("KUBECONFIG2")
			}
			if kubeconfig_d == "" {
				return fmt.Errorf("cannot get destination cluster kube config, please use --destination-kubeconfig or export env KUBECONFIG2 to set")
			}

			// For source
			namespace_s := migrateFlags.SourceNamespace
			client_s, err := p.NewClient(namespace_s)
			if err != nil {
				return err
			}

			err = printServiceWithRevisions(client_s, namespace_s, "source")
			if err != nil {
				return err
			}

			dp := commands.KnParams{
				KubeCfgPath: kubeconfig_d,
			}
			// For destination
			namespace_d := migrateFlags.DestinationNamespace
			dp.Initialize()
			client_d, err := dp.NewClient(namespace_d)
			if err != nil {
				return err
			}

			fmt.Println(color.GreenString("[Before kn-migration]"))
			err = printServiceWithRevisions(client_d, namespace_d, "destination")
			if err != nil {
				return err
			}

			fmt.Println("\nNow kn-migration all Knative resources: \nFrom the source namespace ", color.BlueString(namespace_s), "of cluster", color.CyanString(p.KubeCfgPath))
			fmt.Println("To the destination namespace", color.BlueString(namespace_d), "of cluster", color.CyanString(kubeconfig_d))

			//clientset, err := clientset.NewForConfig(cfg_d)
			//if err != nil {
			//	fmt.Errorf(err.Error())
			//}
			//namespaceExists, err := namespaceExists(*clientset, namespace_d)
			//if err != nil {
			//	fmt.Errorf(err.Error())
			//}
			//
			//if !namespaceExists {
			//	fmt.Println("Create namespace", color.BlueString(namespace_d), "first")
			//	nsSpec := &apiv1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace_d}}
			//	_, err = clientset.CoreV1().Namespaces().Create(nsSpec)
			//}
			//if err != nil {
			//	fmt.Errorf(err.Error())
			//}
			//
			//services_s, err := kn-migration.ListService()
			//if err != nil {
			//	fmt.Errorf(err.Error())
			//}
			//for i := 0; i < len(services_s.Items); i++ {
			//	service_s := services_s.Items[i]
			//	serviceExists, err := kn-migration.ServiceExists(service_s.Name)
			//	if err != nil {
			//		fmt.Errorf(err.Error())
			//	}
			//
			//	if serviceExists {
			//		if cmd.Flag("force").Value.String() == "false" {
			//			fmt.Println("\n[Error] Cannot kn-migration service", color.CyanString(service_s.Name), "in namespace", color.BlueString(namespace_s),
			//				"because the service already exists and no --force option was given")
			//			os.Exit(1)
			//		}
			//		fmt.Println("Deleting service", color.CyanString(service_s.Name), "from the destination cluster first")
			//		kn-migration.DeleteService(service_s.Name)
			//		if err != nil {
			//			fmt.Errorf(err.Error())
			//		}
			//	}
			//
			//	_, err = kn-migration.CreateService(&service_s)
			//	fmt.Println("Migrate service", color.CyanString(service_s.Name), "Successfully")
			//	if err != nil {
			//		fmt.Errorf(err.Error())
			//	}
			//
			//	service_d, err := kn-migration.GetService(service_s.Name)
			//	if err != nil {
			//		fmt.Errorf(err.Error())
			//	}
			//
			//	config, err := kn-migration.GetConfig(service_d.Name)
			//	if err != nil {
			//		fmt.Errorf(err.Error())
			//	}
			//	config_uuid := config.UID
			//
			//	revisions_s, err := kn-migration.ListRevisionByService(service_s.Name)
			//	if err != nil {
			//		fmt.Errorf(err.Error())
			//	}
			//	for i := 0; i < len(revisions_s.Items); i++ {
			//		revision_s := revisions_s.Items[i]
			//		if revision_s.Name != service_s.Status.LatestReadyRevisionName {
			//			revision, err := kn-migration.CreateRevision(&revision_s, config_uuid)
			//			if err != nil {
			//				fmt.Errorf(err.Error())
			//			}
			//			fmt.Println("Migrate revision", color.CyanString(revision.Name), "successfully")
			//		} else {
			//			revision, err := kn-migration.GetRevision(revision_s.Name)
			//			if err != nil {
			//				fmt.Errorf(err.Error())
			//			}
			//			source_revision_generation := revision_s.ObjectMeta.Labels["serving.knative.dev/configurationGeneration"]
			//			revision.ObjectMeta.Labels["serving.knative.dev/configurationGeneration"] = source_revision_generation
			//			_, err = serving_client_d.Revisions(namespace_d).Update(revision)
			//			if err != nil {
			//				fmt.Errorf(err.Error())
			//			}
			//			fmt.Println("Replace revision", color.CyanString(revision_s.Name), "to generation", source_revision_generation, "successfully")
			//		}
			//	}
			//	fmt.Println("")
			//}
			//
			//fmt.Println(color.GreenString("[After kn-migration]"))
			//err = kn-migration.PrintServiceWithRevisions("destination")
			//if err != nil {
			//	fmt.Errorf(err.Error())
			//}
			//
			//if cmd.Flag("delete").Value.String() == "false" {
			//	fmt.Println("Migrate without --delete option, skip deleting Knative resource in source cluster")
			//} else {
			//	fmt.Println("Migrate with --delete option, deleting all Knative resource in source cluster")
			//}
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

	fmt.Println("There are", color.CyanString("%v", len(services.Items)), "service(s) in", clustername, "cluster", color.BlueString(namespace), "namespace:")
	for i := 0; i < len(services.Items); i++ {
		service := services.Items[i]
		color.Cyan("%-25s%-30s%-20s\n", "Name", "Current Revision", "Ready")
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
