package serving

import (
	"fmt"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	client_v1alpha1 "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type NamespacedClient struct {
	client    client_v1alpha1.ServingV1alpha1Interface
	namespace string
}

func NewNamespacedClient(client client_v1alpha1.ServingV1alpha1Interface, namespace string) *NamespacedClient {
	return &NamespacedClient{
		client:    client,
		namespace: namespace,
	}
}

func (cl *NamespacedClient) Service(name string) (*v1alpha1.Service, error) {
	service, err := cl.client.Services(cl.namespace).Get(name, v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	service.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "knative.dev",
		Version: "v1alpha1",
		Kind:    "Service"})
	return service, nil
}

func (cl *NamespacedClient) Revision(name string) (*v1alpha1.Revision, error) {
	revision, err := cl.client.Revisions(cl.namespace).Get(name, v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	updateRevisionWithGVK(revision)
	return revision, nil
}

func (cl *NamespacedClient) Configuration(name string) (*v1alpha1.Configuration, error) {
	cofiguration, err := cl.client.Configurations(cl.namespace).Get(name, v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	cofiguration.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "knative.dev",
		Version: "v1alpha1",
		Kind:    "Configuration"})
	return cofiguration, nil
}

func (cl *NamespacedClient) RevisionsForService(service *v1alpha1.Service) ([]*v1alpha1.Revision, error) {
	labelSelector := fmt.Sprintf("serving.knative.dev/service=%s", service.Name)
	revisions, err := cl.client.Revisions(cl.namespace).List(v1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, err
	}
	revisionSlice := make([]*v1alpha1.Revision, len(revisions.Items))
	for idx, _ := range revisions.Items {
		revision := revisions.Items[idx].DeepCopy()
		updateRevisionWithGVK(revision)
		revisionSlice[idx] = revision
	}
	return revisionSlice, nil
}

// ============================================================================

func updateRevisionWithGVK(revision *v1alpha1.Revision) {
	revision.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "knative.dev",
		Version: "v1alpha1",
		Kind:    "Revision"})
}
