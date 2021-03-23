// Copyright 2020 The Knative Authors

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"knative.dev/client/pkg/wait"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

const (
	ksvcKind = "ksvc"
)

// knServingGitOpsClient - kn service client
// to work on a local repo instead of a remote cluster
type knServingGitOpsClient struct {
	dir        string
	namespace  string
	fileMode   bool
	fileFormat string
	KnServingClient
}

// NewKnServingGitOpsClient returns an instance of the
// kn service gitops client
func NewKnServingGitOpsClient(namespace, dir string) KnServingClient {
	mode, format := getFileModeAndType(dir)
	return &knServingGitOpsClient{
		dir:        dir,
		namespace:  namespace,
		fileMode:   mode,
		fileFormat: format,
	}
}

func (cl *knServingGitOpsClient) getKsvcFilePath(name string) string {
	if cl.fileMode {
		return cl.dir
	}
	return filepath.Join(cl.dir, cl.namespace, ksvcKind, name+".yaml")
}

func getFileModeAndType(dir string) (bool, string) {
	switch {
	case strings.HasSuffix(dir, ".yaml"):
		return true, "yaml"
	case strings.HasSuffix(dir, ".yml"):
		return true, "yaml"
	case strings.HasSuffix(dir, ".json"):
		return true, "json"
	}
	return false, "yaml"
}

// Namespace returns the namespace
func (cl *knServingGitOpsClient) Namespace(context.Context) string {
	return cl.namespace
}

// GetService returns the knative service for the name
func (cl *knServingGitOpsClient) GetService(ctx context.Context, name string) (*servingv1.Service, error) {
	return readServiceFromFile(cl.getKsvcFilePath(name), name)
}

// ListServices lists the services in the path provided
func (cl *knServingGitOpsClient) ListServices(ctx context.Context, opts ...ListConfig) (*servingv1.ServiceList, error) {
	svcs, err := cl.listServicesFromDirectory()
	if err != nil {
		return nil, err
	}
	typeMeta := metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "List",
	}
	serviceList := &servingv1.ServiceList{
		TypeMeta: typeMeta,
		Items:    svcs,
	}
	return serviceList, nil
}

func (cl *knServingGitOpsClient) listServicesFromDirectory() ([]servingv1.Service, error) {
	if cl.fileMode {
		svc, err := readServiceFromFile(cl.dir, "")
		if err != nil {
			return nil, err
		}
		return []servingv1.Service{*svc}, nil
	}
	var services []servingv1.Service
	root := cl.dir
	if cl.namespace != "" {
		root = filepath.Join(cl.dir, cl.namespace)
	}
	if _, err := os.Stat(root); err != nil {
		return nil, err
	}
	if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		switch {
		// skip if dir is not ksvc
		case info.IsDir():
			return nil

		// skip non yaml files
		case !strings.HasSuffix(info.Name(), ".yaml"):
			return nil

		// skip non ksvc dir
		case !strings.Contains(path, ksvcKind):
			return filepath.SkipDir

		default:
			svc, err := readServiceFromFile(path, "")
			if err != nil {
				return err
			}
			services = append(services, *svc)
			return nil
		}
	}); err != nil {
		return nil, err
	}
	return services, nil
}

// CreateService saves the knative service spec in
// yaml format in the local path provided
func (cl *knServingGitOpsClient) CreateService(ctx context.Context, service *servingv1.Service) error {
	updateServingGvk(service)
	if cl.fileMode {
		return writeFile(service, cl.dir, cl.fileFormat)
	}
	//check if dir exist
	if _, err := os.Stat(cl.dir); os.IsNotExist(err) {
		return fmt.Errorf("directory '%s' not present, please create the directory and try again", cl.dir)
	}
	return writeFile(service, cl.getKsvcFilePath(service.ObjectMeta.Name), cl.fileFormat)
}

func writeFile(obj runtime.Object, fp, format string) error {
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		os.MkdirAll(filepath.Dir(fp), 0755)
	}
	w, err := os.Create(fp)
	if err != nil {
		return err
	}
	yamlPrinter, err := genericclioptions.NewJSONYamlPrintFlags().ToPrinter(format)
	if err != nil {
		return err
	}
	return yamlPrinter.PrintObj(obj, w)
}

// UpdateService updates the service in
// the local directory
func (cl *knServingGitOpsClient) UpdateService(ctx context.Context, service *servingv1.Service) (bool, error) {
	// check if file exist
	if _, err := cl.GetService(context.TODO(), service.ObjectMeta.Name); err != nil {
		return false, err
	}
	// replace file
	return true, cl.CreateService(context.TODO(), service)
}

// UpdateServiceWithRetry updates the service in the local directory
func (cl *knServingGitOpsClient) UpdateServiceWithRetry(ctx context.Context, name string, updateFunc ServiceUpdateFunc, nrRetries int) (bool, error) {
	return updateServiceWithRetry(cl, name, updateFunc, nrRetries)
}

// DeleteService removes the file from the local file system
func (cl *knServingGitOpsClient) DeleteService(ctx context.Context, serviceName string, timeout time.Duration) error {
	return os.Remove(cl.getKsvcFilePath(serviceName))
}

// WaitForService always returns success for this client
func (cl *knServingGitOpsClient) WaitForService(ctx context.Context, name string, timeout time.Duration, msgCallback wait.MessageCallback) (error, time.Duration) {
	return nil, 1 * time.Second
}

func readServiceFromFile(fileKey, name string) (*servingv1.Service, error) {
	var svc servingv1.Service
	file, err := os.Open(fileKey)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, apierrors.NewNotFound(servingv1.Resource("services"), name)
		}
		return nil, err
	}
	decoder := yaml.NewYAMLOrJSONDecoder(file, 512)
	if err := decoder.Decode(&svc); err != nil {
		return nil, err
	}
	return &svc, nil
}
