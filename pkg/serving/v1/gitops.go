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

	"k8s.io/apimachinery/pkg/watch"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

const (
	operationNotSuportedError = "this operation is not supported in gitops mode"
	ksvcKind                  = "ksvc"
)

// knServingGitOpsClient - kn service client
// to work on a local repo instead of a remote cluster
type knServingGitOpsClient struct {
	dir        string
	namespace  string
	fileClient fileOpsClient
}

type fileOpsClient interface {
	writeToFile(runtime.Object, string) error
	readFromFile(string, string) (*servingv1.Service, error)
	removeFile(string) error
	listFiles(string) ([]servingv1.Service, error)
}

type fileClient struct{}

// NewKnServingGitOpsClient returns an instance of the
// kn service gitops client
func NewKnServingGitOpsClient(namespace, dir string) *knServingGitOpsClient {
	// create dir , if not present
	namespaceDir := filepath.Join(dir, namespace, ksvcKind)
	if _, err := os.Stat(namespaceDir); os.IsNotExist(err) {
		os.MkdirAll(namespaceDir, 0777)
	}
	return &knServingGitOpsClient{
		dir:        dir,
		namespace:  namespace,
		fileClient: &fileClient{},
	}
}

func (f *fileClient) readFromFile(fileKey, name string) (*servingv1.Service, error) {
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
	// updateServingGvk(&svc)
	return &svc, nil
}

func (f *fileClient) writeToFile(obj runtime.Object, filePath string) error {
	objFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	yamlPrinter, err := genericclioptions.NewJSONYamlPrintFlags().ToPrinter("yaml")
	if err != nil {
		return err
	}
	return yamlPrinter.PrintObj(obj, objFile)
}

func (f *fileClient) removeFile(filePath string) error {
	return os.Remove(filePath)
}

// ListServices lists the services in the path provided
func (f *fileClient) listFiles(root string) ([]servingv1.Service, error) {
	var services []servingv1.Service

	if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		switch {
		// skip if dir is not ksvc
		case info.IsDir():
			return nil

		// skip non yaml files
		case !strings.Contains(info.Name(), ".yaml"):
			return nil

		// skip non ksvc dir
		case !strings.Contains(path, ksvcKind):
			return filepath.SkipDir

		default:
			svc, err := f.readFromFile(path, "")
			if err != nil {
				return err
			}
			// updateServingGvk(svc)
			services = append(services, *svc)
			return nil
		}
	}); err != nil {
		return nil, err
	}
	return services, nil
}

func (cl *knServingGitOpsClient) getKsvcFilePath(name string) string {
	return filepath.Join(cl.dir, cl.namespace, ksvcKind, name+".yaml")
}

// Namespace returns the namespace
func (cl *knServingGitOpsClient) Namespace() string {
	return cl.namespace
}

// GetService returns the knative service for the name
func (cl *knServingGitOpsClient) GetService(name string) (*servingv1.Service, error) {
	return cl.fileClient.readFromFile(cl.getKsvcFilePath(name), name)
}

// WatchService is not supported by this client
func (cl *knServingGitOpsClient) WatchService(name string, timeout time.Duration) (watch.Interface, error) {
	return nil, fmt.Errorf(operationNotSuportedError)
}

// WatchRevision is not supported by this client
func (cl *knServingGitOpsClient) WatchRevision(name string, timeout time.Duration) (watch.Interface, error) {
	return nil, fmt.Errorf(operationNotSuportedError)
}

// ListServices lists the services in the path provided
func (cl *knServingGitOpsClient) ListServices(config ...ListConfig) (*servingv1.ServiceList, error) {
	var root string

	switch cl.namespace {
	case "":
		root = cl.dir
	default:
		root = filepath.Join(cl.dir, cl.namespace)
	}

	services, err := cl.fileClient.listFiles(root)
	if err != nil {
		return nil, err
	}

	typeMeta := metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "List",
	}
	serviceList := &servingv1.ServiceList{
		TypeMeta: typeMeta,
		Items:    services,
	}

	return serviceList, nil
}

// CreateService saves the knative service spec in
// yaml format in the local path provided
func (cl *knServingGitOpsClient) CreateService(service *servingv1.Service) error {
	updateServingGvk(service)
	return cl.fileClient.writeToFile(service, cl.getKsvcFilePath(service.ObjectMeta.Name))
}

// UpdateService updates the service in
// the local directory
func (cl *knServingGitOpsClient) UpdateService(service *servingv1.Service) error {
	// check if file exist
	if _, err := cl.GetService(service.ObjectMeta.Name); err != nil {
		return err
	}
	// replace file
	return cl.CreateService(service)
}

// UpdateServiceWithRetry updates the service in the local directory
func (cl *knServingGitOpsClient) UpdateServiceWithRetry(name string, updateFunc ServiceUpdateFunc, nrRetries int) error {
	return updateServiceWithRetry(cl, name, updateFunc, nrRetries)
}

// ApplyService is not supported by this client
func (cl *knServingGitOpsClient) ApplyService(modifiedService *servingv1.Service) (bool, error) {
	return false, fmt.Errorf(operationNotSuportedError)
}

// DeleteService removes the file from the local file system
func (cl *knServingGitOpsClient) DeleteService(serviceName string, timeout time.Duration) error {
	return cl.fileClient.removeFile(cl.getKsvcFilePath(serviceName))
}

// WaitForService always returns success for this client
func (cl *knServingGitOpsClient) WaitForService(name string, timeout time.Duration, msgCallback wait.MessageCallback) (error, time.Duration) {
	return nil, 1 * time.Second
}

// GetConfiguration not supported by this client
func (cl *knServingGitOpsClient) GetConfiguration(name string) (*servingv1.Configuration, error) {
	return nil, fmt.Errorf(operationNotSuportedError)
}

// GetRevision not supported by this client
func (cl *knServingGitOpsClient) GetRevision(name string) (*servingv1.Revision, error) {
	return nil, fmt.Errorf(operationNotSuportedError)
}

// GetBaseRevision not supported by this client
func (cl *knServingGitOpsClient) GetBaseRevision(service *servingv1.Service) (*servingv1.Revision, error) {
	return nil, fmt.Errorf(operationNotSuportedError)
}

// CreateRevision not supported by this client
func (cl *knServingGitOpsClient) CreateRevision(revision *servingv1.Revision) error {
	return fmt.Errorf(operationNotSuportedError)
}

// UpdateRevision not supported by this client
func (cl *knServingGitOpsClient) UpdateRevision(revision *servingv1.Revision) error {
	return fmt.Errorf(operationNotSuportedError)
}

// DeleteRevision not supported by this client
func (cl *knServingGitOpsClient) DeleteRevision(name string, timeout time.Duration) error {
	return fmt.Errorf(operationNotSuportedError)
}

// ListRevisions not supported by this client
func (cl *knServingGitOpsClient) ListRevisions(config ...ListConfig) (*servingv1.RevisionList, error) {
	return nil, fmt.Errorf(operationNotSuportedError)
}

// GetRoute not supported by this client
func (cl *knServingGitOpsClient) GetRoute(name string) (*servingv1.Route, error) {
	return nil, fmt.Errorf(operationNotSuportedError)
}

// ListRoutes not supported by this client
func (cl *knServingGitOpsClient) ListRoutes(config ...ListConfig) (*servingv1.RouteList, error) {
	return nil, fmt.Errorf(operationNotSuportedError)
}
