// Copyright © 2018 The Knative Authors
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
	"strings"
	"testing"

	"gotest.tools/assert"
	"k8s.io/apimachinery/pkg/runtime"
	client_testing "k8s.io/client-go/testing"
	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/util"
	"knative.dev/serving/pkg/apis/serving/v1alpha1"
	"knative.dev/serving/pkg/apis/serving/v1beta1"
)

func fakeRouteList(args []string, response *v1alpha1.RouteList) (action client_testing.Action, output []string, err error) {
	knParams := &commands.KnParams{}
	cmd, fakeServing, buf := commands.CreateTestKnCommand(NewRouteCommand(knParams), knParams)
	fakeServing.AddReactor("*", "*",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			action = a
			return true, response, nil
		})
	cmd.SetArgs(args)
	err = cmd.Execute()
	if err != nil {
		return
	}
	output = strings.Split(buf.String(), "\n")
	return
}

func TestListEmpty(t *testing.T) {
	action, output, err := fakeRouteList([]string{"route", "list"}, &v1alpha1.RouteList{})
	if err != nil {
		t.Error(err)
		return
	}
	if action == nil {
		t.Errorf("No action")
	} else if !action.Matches("list", "routes") {
		t.Errorf("Bad action %v", action)
	} else if output[0] != "No resources found." {
		t.Errorf("Bad output %s", output[0])
	}
}

func TestRouteListDefaultOutput(t *testing.T) {
	route1 := createMockRouteSingleTarget("foo", "foo-01234", 100)
	route2 := createMockRouteSingleTarget("bar", "bar-98765", 100)
	routeList := &v1alpha1.RouteList{Items: []v1alpha1.Route{*route1, *route2}}
	action, output, err := fakeRouteList([]string{"route", "list"}, routeList)
	if err != nil {
		t.Fatal(err)
	}
	if action == nil {
		t.Errorf("No action")
	} else if !action.Matches("list", "routes") {
		t.Errorf("Bad action %v", action)
	}
	assert.Check(t, util.ContainsAll(output[0], "NAME", "URL", "AGE", "CONDITIONS", "TRAFFIC"))
	assert.Check(t, util.ContainsAll(output[1], "foo", "100% -> foo-01234"))
	assert.Check(t, util.ContainsAll(output[2], "bar", "100% -> bar-98765"))
}

func TestRouteListWithTwoTargetsOutput(t *testing.T) {
	route := createMockRouteTwoTarget("foo", "foo-01234", "foo-98765", 20, 80)
	routeList := &v1alpha1.RouteList{Items: []v1alpha1.Route{*route}}
	action, output, err := fakeRouteList([]string{"route", "list"}, routeList)
	if err != nil {
		t.Fatal(err)
	}
	if action == nil {
		t.Errorf("No action")
	} else if !action.Matches("list", "routes") {
		t.Errorf("Bad action %v", action)
	}
	assert.Check(t, util.ContainsAll(output[0], "NAME", "URL", "AGE", "CONDITIONS", "TRAFFIC"))
	assert.Check(t, util.ContainsAll(output[1], "foo", "20% -> foo-01234, 80% -> foo-98765"))
}

func createMockRouteMeta(name string) *v1alpha1.Route {
	route := &v1alpha1.Route{}
	route.Kind = "Route"
	route.APIVersion = "knative.dev/v1alpha1"
	route.Name = name
	route.Namespace = commands.FakeNamespace
	return route
}

func createMockTrafficTarget(revision string, percent int) *v1alpha1.TrafficTarget {
	return &v1alpha1.TrafficTarget{
		TrafficTarget: v1beta1.TrafficTarget{
			RevisionName: revision,
			Percent:      percent,
		},
	}
}

func createMockRouteSingleTarget(name, revision string, percent int) *v1alpha1.Route {
	route := createMockRouteMeta(name)
	target := createMockTrafficTarget(revision, percent)
	route.Status.Traffic = []v1alpha1.TrafficTarget{*target}
	return route
}

func createMockRouteTwoTarget(name string, rev1, rev2 string, percent1, percent2 int) *v1alpha1.Route {
	route := createMockRouteMeta(name)
	target1 := createMockTrafficTarget(rev1, percent1)
	target2 := createMockTrafficTarget(rev2, percent2)
	route.Status.Traffic = []v1alpha1.TrafficTarget{*target1, *target2}
	return route
}
