/*
 Copyright 2024 The Knative Authors

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package tui_test

import (
	"testing"

	"knative.dev/client/pkg/context"
	"knative.dev/client/pkg/output/tui"
)

func TestNewWidgets(t *testing.T) {
	t.Parallel()
	ctx := context.TestContext(t)
	w := tui.NewWidgets(ctx)

	if w == nil {
		t.Errorf("want widgets, got nil")
	}
}

func TestNewInteractiveWidgets(t *testing.T) {
	t.Parallel()
	ctx := context.TestContext(t)
	w, err := tui.NewInteractiveWidgets(ctx)

	if err == nil {
		t.Error("want error, got nil")
	}
	if w != nil {
		t.Errorf("want nil, got %v", w)
	}
}
