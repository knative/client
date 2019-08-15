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

package route

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/knative/client/pkg/kn/commands"
	"gotest.tools/assert"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	client_testing "k8s.io/client-go/testing"
	"knative.dev/serving/pkg/apis/serving/v1alpha1"
	"sigs.k8s.io/yaml"
)

func fakeRouteDescribe(args []string, response *v1alpha1.Route) (action client_testing.Action, output string, err error) {
	knParams := &commands.KnParams{}
	cmd, fakeRoute, buf := commands.CreateTestKnCommand(NewRouteCommand(knParams), knParams)
	fakeRoute.AddReactor("*", "*",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			action = a
			return true, response, nil
		})
	cmd.SetArgs(args)
	err = cmd.Execute()
	if err != nil {
		return
	}
	output = buf.String()
	return
}

func TestCompletion(t *testing.T) {
	var expectedRoute v1alpha1.Route

	setup := func(t *testing.T) {
		expectedRoute = v1alpha1.Route{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Route",
				APIVersion: "knative.dev/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "default",
			},
		}
	}

	t.Run("requires the route name", func(t *testing.T) {
		_, _, err := fakeRouteDescribe([]string{"route", "describe"}, &v1alpha1.Route{})
		assert.Assert(t, err != nil)
		assert.Assert(t, strings.Contains(err.Error(), "requires the route name."))
	})

	t.Run("describe a valid route with default output", func(t *testing.T) {
		setup(t)

		action, output, err := fakeRouteDescribe([]string{"route", "describe", "foo"}, &expectedRoute)
		assert.Assert(t, err == nil)
		assert.Assert(t, action != nil)
		assert.Assert(t, action.Matches("get", "routes"))

		jsonData, err := yaml.YAMLToJSON([]byte(output))
		assert.Assert(t, err == nil)

		var returnedRoute v1alpha1.Route
		err = json.Unmarshal(jsonData, &returnedRoute)
		assert.Assert(t, err == nil)
		assert.Assert(t, equality.Semantic.DeepEqual(expectedRoute, returnedRoute))
	})

	t.Run("describe a valid route with special output", func(t *testing.T) {
		t.Run("yaml", func(t *testing.T) {
			setup(t)

			action, output, err := fakeRouteDescribe([]string{"route", "describe", "foo", "-oyaml"}, &expectedRoute)
			assert.Assert(t, err == nil)
			assert.Assert(t, action != nil)
			assert.Assert(t, action.Matches("get", "routes"))

			jsonData, err := yaml.YAMLToJSON([]byte(output))
			assert.Assert(t, err == nil)

			var returnedRoute v1alpha1.Route
			err = json.Unmarshal(jsonData, &returnedRoute)
			assert.Assert(t, err == nil)
			assert.Assert(t, equality.Semantic.DeepEqual(expectedRoute, returnedRoute))
		})

		t.Run("json", func(t *testing.T) {
			setup(t)

			action, output, err := fakeRouteDescribe([]string{"route", "describe", "foo", "-ojson"}, &expectedRoute)
			assert.Assert(t, err == nil)
			assert.Assert(t, action != nil)
			assert.Assert(t, action.Matches("get", "routes"))

			var returnedRoute v1alpha1.Route
			err = json.Unmarshal([]byte(output), &returnedRoute)
			assert.Assert(t, err == nil)
			assert.Assert(t, equality.Semantic.DeepEqual(expectedRoute, returnedRoute))
		})

		t.Run("name", func(t *testing.T) {
			setup(t)

			action, output, err := fakeRouteDescribe([]string{"route", "describe", "foo", "-oname"}, &expectedRoute)
			assert.Assert(t, err == nil)
			assert.Assert(t, action != nil)
			assert.Assert(t, action.Matches("get", "routes"))
			assert.Assert(t, strings.Contains(output, expectedRoute.Name))
		})
	})
}
