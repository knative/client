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
	"knative.dev/pkg/ptr"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/util"
)

func fakeRouteList(args []string, response *servingv1.RouteList) (action client_testing.Action, output []string, err error) {
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
	action, output, err := fakeRouteList([]string{"route", "list"}, &servingv1.RouteList{})
	assert.NilError(t, err)
	if action == nil {
		t.Errorf("No action")
	} else if !action.Matches("list", "routes") {
		t.Errorf("Bad action %v", action)
	} else if output[0] != "No routes found." {
		t.Errorf("Bad output %s", output[0])
	}
}

func TestRouteListDefaultOutput(t *testing.T) {
	route1 := createMockRouteSingleTarget("foo", "foo-01234", 100)
	route2 := createMockRouteSingleTarget("bar", "bar-98765", 100)
	routeList := &servingv1.RouteList{Items: []servingv1.Route{*route1, *route2}}
	action, output, err := fakeRouteList([]string{"route", "list"}, routeList)
	assert.NilError(t, err)
	if action == nil {
		t.Errorf("No action")
	} else if !action.Matches("list", "routes") {
		t.Errorf("Bad action %v", action)
	}
	assert.Check(t, util.ContainsAll(output[0], "NAME", "URL", "READY"))
	assert.Check(t, util.ContainsAll(output[1], "foo"))
	assert.Check(t, util.ContainsAll(output[2], "bar"))
}

func TestRouteListDefaultOutputNoHeaders(t *testing.T) {
	route1 := createMockRouteSingleTarget("foo", "foo-01234", 100)
	route2 := createMockRouteSingleTarget("bar", "bar-98765", 100)
	routeList := &servingv1.RouteList{Items: []servingv1.Route{*route1, *route2}}
	action, output, err := fakeRouteList([]string{"route", "list", "--no-headers"}, routeList)
	assert.NilError(t, err)
	if action == nil {
		t.Errorf("No action")
	} else if !action.Matches("list", "routes") {
		t.Errorf("Bad action %v", action)
	}

	assert.Check(t, util.ContainsNone(output[0], "NAME", "URL", "READY"))
	assert.Check(t, util.ContainsAll(output[0], "foo"))
	assert.Check(t, util.ContainsAll(output[1], "bar"))

}

func TestRouteListWithTwoTargetsOutput(t *testing.T) {
	route := createMockRouteTwoTarget("foo", "foo-01234", "foo-98765", 20, 80)
	routeList := &servingv1.RouteList{Items: []servingv1.Route{*route}}
	action, output, err := fakeRouteList([]string{"route", "list"}, routeList)
	assert.NilError(t, err)
	if action == nil {
		t.Errorf("No action")
	} else if !action.Matches("list", "routes") {
		t.Errorf("Bad action %v", action)
	}
	assert.Check(t, util.ContainsAll(output[0], "NAME", "URL", "READY"))
	assert.Check(t, util.ContainsAll(output[1], "foo"))
}

func createMockRouteMeta(name string) *servingv1.Route {
	route := &servingv1.Route{}
	route.Kind = "Route"
	route.APIVersion = "serving.knative.dev/v1"
	route.Name = name
	route.Namespace = commands.FakeNamespace
	return route
}

func createMockTrafficTarget(revision string, percent int) *servingv1.TrafficTarget {
	return &servingv1.TrafficTarget{
		RevisionName: revision,
		Percent:      ptr.Int64(int64(percent)),
	}
}

func createMockRouteSingleTarget(name, revision string, percent int) *servingv1.Route {
	route := createMockRouteMeta(name)
	target := createMockTrafficTarget(revision, percent)
	route.Status.Traffic = []servingv1.TrafficTarget{*target}
	return route
}

func createMockRouteTwoTarget(name string, rev1, rev2 string, percent1, percent2 int) *servingv1.Route {
	route := createMockRouteMeta(name)
	target1 := createMockTrafficTarget(rev1, percent1)
	target2 := createMockTrafficTarget(rev2, percent2)
	route.Status.Traffic = []servingv1.TrafficTarget{*target1, *target2}
	return route
}
