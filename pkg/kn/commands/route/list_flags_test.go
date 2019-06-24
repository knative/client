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
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or im
// See the License for the specific language governing permissions and
// limitations under the License.

package route

import (
	"reflect"
	"testing"

	"github.com/knative/client/pkg/kn/commands"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func TestRoutListFlags(t *testing.T) {
	testObject := createMockRouteMeta("foo")
	knParams := &commands.KnParams{}
	cmd, _, buf := commands.CreateTestKnCommand(NewRouteCommand(knParams), knParams)
	routeListFlags := NewRouteListFlags()
	routeListFlags.AddFlags(cmd)
	printer, err := routeListFlags.ToPrinter()
	if genericclioptions.IsNoCompatiblePrinterError(err) {
		t.Fatalf("Expected to match human readable printer.")
	}
	if err != nil {
		t.Fatalf("Failed to find a proper printer.")
	}
	err = printer.PrintObj(testObject, buf)
	if err != nil {
		t.Fatalf("Failed to print the object.")
	}
	actualFormats := routeListFlags.AllowedFormats()
	expectedFormats := []string{"json", "yaml", "name", "go-template", "go-template-file", "template", "templatefile", "jsonpath", "jsonpath-file"}
	if reflect.DeepEqual(actualFormats, expectedFormats) {
		t.Fatalf("Expecting allowed formats:\n%s\nFound:\n%s\n", expectedFormats, actualFormats)
	}
}
