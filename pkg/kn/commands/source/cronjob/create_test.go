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

package cronjob

import (
	"errors"
	"fmt"
	"testing"

	"gotest.tools/assert"
	"k8s.io/apimachinery/pkg/runtime"
	client_testing "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/clientcmd"
	"knative.dev/eventing/pkg/apis/sources/v1alpha1"
	serving_v1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"

	sources_fake "knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1alpha1/fake"

	v1alpha12 "knative.dev/client/pkg/eventing/sources/v1alpha1"
	knservingclient "knative.dev/client/pkg/serving/v1alpha1"
	"knative.dev/client/pkg/util"
)

func TestCreateCronJobSource(t *testing.T) {

	servingClient := knservingclient.NewMockKnServiceClient(t)

	recorder := servingClient.Recorder()
	recorder.GetService("mysvc", &serving_v1alpha1.Service{}, nil)

	sources := prepareFakeCronJobSourceClientFactory()
	sources.AddReactor("create", "cronjobsources",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			createAction, ok := a.(client_testing.CreateAction)
			if !ok {
				return true, nil, fmt.Errorf("wrong kind of action %v", a)
			}
			src, ok := createAction.GetObject().(*v1alpha1.CronJobSource)
			if !ok {
				return true, nil, errors.New("was passed the wrong object")
			}
			return true, src, nil
		})

	out, err := executeCronJobSourceCommand(servingClient, "create", "--sink", "svc:mysvc", "--schedule", "* * * * *", "testsource")
	assert.NilError(t, err, "Source should have been created")
	util.ContainsAll(out, "created", "default", "testsource")
}

func prepareFakeCronJobSourceClientFactory() *sources_fake.FakeSourcesV1alpha1 {
	sources := &sources_fake.FakeSourcesV1alpha1{&client_testing.Fake{}}
	cronJobSourceClientFactory = func(config clientcmd.ClientConfig, namespace string) (client v1alpha12.KnCronJobSourcesClient, err error) {
		return v1alpha12.NewKnSourcesClient(sources, namespace).CronJobSourcesClient(), nil
	}
	return sources
}
