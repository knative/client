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

package errors

import (
	"github.com/pkg/errors"
	"gotest.tools/assert"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"testing"
)

func TestBuild(t *testing.T) {
	//default non api error
	defaultError := errors.New("my-custom-error")
	err := Build(defaultError)
	assert.Error(t, err, "my-custom-error")

	gv := schema.GroupResource{
		Group: "serving.knative.dev",
		Resource: "service",
	}

	//api error containing expected error when knative crd is not available
	apiError := api_errors.NewNotFound(gv, "serv")
	apiError.Status().Details.Causes = []v1.StatusCause{
		{
			Type: "UnexpectedServerResponse",
			Message: "404 page not found",
		},
	}
	err = Build(apiError)
	assert.Error(t, err, "no Knative serving API found on the backend. Please verify the installation.")

	//api error not registered in error factory
	apiError = api_errors.NewAlreadyExists(gv, "serv")
	err = Build(apiError)
	assert.Error(t, err, "service.serving.knative.dev \"serv\" already exists")

	//default not found api error
	apiError = api_errors.NewNotFound(gv, "serv")
	err = Build(apiError)
	assert.Error(t, err, "service.serving.knative.dev \"serv\" not found")
}
