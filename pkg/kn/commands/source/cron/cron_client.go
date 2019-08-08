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

package cron

import (
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"knative.dev/eventing/pkg/apis/sources/v1alpha1"
	sources_v1alpha1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1alpha1"
)

type CronSourceClient struct {
	client    *sources_v1alpha1.SourcesV1alpha1Client
	namespace string
}

func NewCronSourceClient(config clientcmd.ClientConfig, namespace string) (*CronSourceClient, error) {
	clientConfig, err := config.ClientConfig()
	if err != nil {
		return nil, err
	}
	client, err := sources_v1alpha1.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}
	return &CronSourceClient{
		namespace: namespace,
		client:    client,
	}, nil
}

func (c *CronSourceClient) CreateCronSource(name, schedule, data string, sink *v1.ObjectReference) error {
	source := v1alpha1.CronJobSource{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.CronJobSourceSpec{
			Schedule: schedule,
			Data:     data,
		},
	}
	if sink != nil {
		source.Spec.Sink = sink
	}
	_, err := c.client.CronJobSources(c.namespace).Create(&source)
	return err
}
