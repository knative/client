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

package test

import (
	"context"
	"encoding/json"
	"strings"

	"gotest.tools/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientv1alpha1 "knative.dev/client/pkg/apis/client/v1alpha1"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/ptr"
	pkgtest "knative.dev/pkg/test"
	"knative.dev/serving/pkg/apis/config"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	servingtest "knative.dev/serving/pkg/testing/v1"

	"knative.dev/client/pkg/util"
)

// ExpectedServiceListOption enables further configuration of a ServiceList.
type ExpectedServiceListOption func(*servingv1.ServiceList)

// ExpectedRevisionListOption enables further configuration of a RevisionList.
type ExpectedRevisionListOption func(*servingv1.RevisionList)

// ExpectedKNExportOption enables further configuration of a Export.
type ExpectedKNExportOption func(*clientv1alpha1.Export)

// ServiceCreate verifies given service creation in sync mode and also verifies output
func ServiceCreate(r *KnRunResultCollector, serviceName string) {
	out := r.KnTest().Kn().Run("service", "create", serviceName, "--image", pkgtest.ImagePath("helloworld"))
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "service", serviceName, "creating", "namespace", r.KnTest().Kn().Namespace(), "ready"))
}

// ServiceListEmpty verifies that there are no services present
func ServiceListEmpty(r *KnRunResultCollector) {
	out := r.KnTest().Kn().Run("service", "list")
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stdout, "No services found."))
}

// ServiceList verifies if given service exists
func ServiceList(r *KnRunResultCollector, serviceName string) {
	out := r.KnTest().Kn().Run("service", "list", serviceName)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stdout, serviceName))
}

// ServiceDescribe describes given service and verifies the keys in the output
func ServiceDescribe(r *KnRunResultCollector, serviceName string) {
	out := r.KnTest().Kn().Run("service", "describe", serviceName)
	r.AssertNoError(out)
	assert.Assert(r.T(), util.ContainsAll(out.Stdout, serviceName, r.KnTest().Kn().Namespace(), pkgtest.ImagePath("helloworld")))
	assert.Assert(r.T(), util.ContainsAll(out.Stdout, "Conditions", "ConfigurationsReady", "Ready", "RoutesReady"))
	assert.Assert(r.T(), util.ContainsAll(out.Stdout, "Name", "Namespace", "URL", "Age", "Revisions"))
}

// ServiceListOutput verifies listing given service using '--output name' flag
func ServiceListOutput(r *KnRunResultCollector, serviceName string) {
	out := r.KnTest().Kn().Run("service", "list", serviceName, "--output", "name")
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stdout, serviceName, "service.serving.knative.dev"))
}

// ServiceUpdate verifies service update operation with given arguments in sync mode
func ServiceUpdate(r *KnRunResultCollector, serviceName string, args ...string) {
	fullArgs := append([]string{}, "service", "update", serviceName)
	fullArgs = append(fullArgs, args...)
	out := r.KnTest().Kn().Run(fullArgs...)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "updating", "service", serviceName, "ready"))
}

// ServiceUpdateWithError verifies service update operation with given arguments in sync mode
// when expecting an error
func ServiceUpdateWithError(r *KnRunResultCollector, serviceName string, args ...string) {
	fullArgs := append([]string{}, "service", "update", serviceName)
	fullArgs = append(fullArgs, args...)
	out := r.KnTest().Kn().Run(fullArgs...)
	r.AssertError(out)
}

// ServiceDelete verifies service deletion in sync mode
func ServiceDelete(r *KnRunResultCollector, serviceName string) {
	out := r.KnTest().Kn().Run("service", "delete", "--wait", serviceName)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stdout, "Service", serviceName, "successfully deleted in namespace", r.KnTest().Kn().Namespace()))
}

// ServiceDescribeWithJSONPath returns output of given JSON path by describing the service
func ServiceDescribeWithJSONPath(r *KnRunResultCollector, serviceName, jsonpath string) string {
	out := r.KnTest().Kn().Run("service", "describe", serviceName, "-o", jsonpath)
	r.AssertNoError(out)
	return out.Stdout
}

// ValidateServiceResources validates cpu and mem resources
func ValidateServiceResources(r *KnRunResultCollector, serviceName string, requestsMemory, requestsCPU, limitsMemory, limitsCPU string) {
	var err error
	rlist := corev1.ResourceList{}
	rlist[corev1.ResourceCPU], err = resource.ParseQuantity(requestsCPU)
	assert.NilError(r.T(), err)
	rlist[corev1.ResourceMemory], err = resource.ParseQuantity(requestsMemory)
	assert.NilError(r.T(), err)

	llist := corev1.ResourceList{}
	llist[corev1.ResourceCPU], err = resource.ParseQuantity(limitsCPU)
	assert.NilError(r.T(), err)
	llist[corev1.ResourceMemory], err = resource.ParseQuantity(limitsMemory)
	assert.NilError(r.T(), err)

	out := r.KnTest().Kn().Run("service", "describe", serviceName, "-ojson")
	data := json.NewDecoder(strings.NewReader(out.Stdout))
	var service servingv1.Service
	err = data.Decode(&service)
	assert.NilError(r.T(), err)

	serviceRequestResourceList := service.Spec.Template.Spec.Containers[0].Resources.Requests
	assert.DeepEqual(r.T(), serviceRequestResourceList, rlist)

	serviceLimitsResourceList := service.Spec.Template.Spec.Containers[0].Resources.Limits
	assert.DeepEqual(r.T(), serviceLimitsResourceList, llist)
}

// GetServiceFromKNServiceDescribe runs the kn service describe command
// decodes it into a ksvc and returns it.
func GetServiceFromKNServiceDescribe(r *KnRunResultCollector, serviceName string) servingv1.Service {
	out := r.KnTest().Kn().Run("service", "describe", serviceName, "-ojson")
	data := json.NewDecoder(strings.NewReader(out.Stdout))
	data.UseNumber()
	var service servingv1.Service
	err := data.Decode(&service)
	assert.NilError(r.T(), err)
	return service
}

// BuildServiceListWithOptions returns ServiceList with options provided
func BuildServiceListWithOptions(options ...ExpectedServiceListOption) *servingv1.ServiceList {
	list := &servingv1.ServiceList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "List",
		},
	}

	for _, fn := range options {
		fn(list)
	}

	return list
}

// WithService appends the given service to ServiceList
func WithService(svc *servingv1.Service) ExpectedServiceListOption {
	return func(list *servingv1.ServiceList) {
		list.Items = append(list.Items, *svc)
	}
}

// BuildRevisionListWithOptions returns RevisionList with options provided
func BuildRevisionListWithOptions(options ...ExpectedRevisionListOption) *servingv1.RevisionList {
	list := &servingv1.RevisionList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "List",
		},
	}

	for _, fn := range options {
		fn(list)
	}

	return list
}

// BuildKNExportWithOptions returns Export object with the options provided
func BuildKNExportWithOptions(options ...ExpectedKNExportOption) *clientv1alpha1.Export {
	knExport := &clientv1alpha1.Export{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "client.knative.dev/v1alpha1",
			Kind:       "Export",
		},
	}

	for _, fn := range options {
		fn(knExport)
	}

	return knExport
}

// BuildConfigurationSpec builds servingv1.ConfigurationSpec with the options provided
func BuildConfigurationSpec(co ...servingtest.ConfigOption) *servingv1.ConfigurationSpec {
	c := &servingv1.Configuration{
		Spec: servingv1.ConfigurationSpec{
			Template: servingv1.RevisionTemplateSpec{
				Spec: *BuildRevisionSpec(pkgtest.ImagePath("helloworld")),
			},
		},
	}
	for _, opt := range co {
		opt(c)
	}
	c.SetDefaults(context.Background())
	return &c.Spec
}

// BuildRevisionSpec for provided image
func BuildRevisionSpec(image string) *servingv1.RevisionSpec {
	return &servingv1.RevisionSpec{
		PodSpec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Image: image,
			}},
			EnableServiceLinks: ptr.Bool(false),
		},
		TimeoutSeconds: ptr.Int64(config.DefaultRevisionTimeoutSeconds),
	}
}

// BuildServiceWithOptions returns ksvc with options provided
func BuildServiceWithOptions(name string, so ...servingtest.ServiceOption) *servingv1.Service {
	svc := servingtest.ServiceWithoutNamespace(name, so...)
	svc.TypeMeta = metav1.TypeMeta{
		Kind:       "Service",
		APIVersion: "serving.knative.dev/v1",
	}
	svc.Spec.Template.Spec.Containers[0].Resources = corev1.ResourceRequirements{}
	return svc
}

// WithTrafficSpec adds route to ksvc
func WithTrafficSpec(revisions []string, percentages []int, tags []string) servingtest.ServiceOption {
	return func(svc *servingv1.Service) {
		var trafficTargets []servingv1.TrafficTarget
		for i, rev := range revisions {
			trafficTargets = append(trafficTargets, servingv1.TrafficTarget{
				Percent: ptr.Int64(int64(percentages[i])),
			})
			if tags[i] != "" {
				trafficTargets[i].Tag = tags[i]
			}
			if rev == "latest" {
				trafficTargets[i].LatestRevision = ptr.Bool(true)
			} else {
				trafficTargets[i].RevisionName = rev
				trafficTargets[i].LatestRevision = ptr.Bool(false)
			}
		}
		svc.Spec.RouteSpec = servingv1.RouteSpec{
			Traffic: trafficTargets,
		}
	}
}

// BuildRevision returns Revision object with the options provided
func BuildRevision(name string, options ...servingtest.RevisionOption) *servingv1.Revision {
	rev := servingtest.Revision("", name, options...)
	rev.TypeMeta = metav1.TypeMeta{
		Kind:       "Revision",
		APIVersion: "serving.knative.dev/v1",
	}
	rev.Spec.PodSpec.Containers[0].Name = config.DefaultUserContainerName
	rev.Spec.PodSpec.EnableServiceLinks = ptr.Bool(false)
	rev.ObjectMeta.SelfLink = ""
	rev.ObjectMeta.Namespace = ""
	rev.ObjectMeta.UID = ""
	rev.ObjectMeta.Generation = int64(0)
	rev.Spec.PodSpec.Containers[0].Resources = corev1.ResourceRequirements{}
	return rev
}

// WithRevision appends Revision object to RevisionList
func WithRevision(rev servingv1.Revision) ExpectedRevisionListOption {
	return func(list *servingv1.RevisionList) {
		list.Items = append(list.Items, rev)
	}
}

// WithKNRevision appends Revision object RevisionList to Kn Export
func WithKNRevision(rev servingv1.Revision) ExpectedKNExportOption {
	return func(export *clientv1alpha1.Export) {
		export.Spec.Revisions = append(export.Spec.Revisions, rev)
	}
}

// WithRevisionEnv adds env variable to Revision object
func WithRevisionEnv(evs ...corev1.EnvVar) servingtest.RevisionOption {
	return func(s *servingv1.Revision) {
		s.Spec.PodSpec.Containers[0].Env = evs
	}
}

// WithRevisionImage adds revision image to Revision object
func WithRevisionImage(image string) servingtest.RevisionOption {
	return func(s *servingv1.Revision) {
		s.Spec.PodSpec.Containers[0].Image = image
	}
}

// WithRevisionAnnotations adds annotation to revision spec in ksvc
func WithRevisionAnnotations(annotations map[string]string) servingtest.ServiceOption {
	return func(service *servingv1.Service) {
		service.Spec.Template.ObjectMeta.Annotations = kmeta.UnionMaps(
			service.Spec.Template.ObjectMeta.Annotations, annotations)
	}
}
