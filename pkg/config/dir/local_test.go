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

package dir_test

import (
	"strings"
	"testing"

	"knative.dev/client/pkg/config/dir"
	"knative.dev/client/pkg/context"
)

func TestConfig(t *testing.T) {
	ctx := context.TestContext(t)
	want := t.TempDir()
	ctx = dir.WithConfigDir(ctx, want)

	got := dir.Config(ctx)
	if got != want {
		t.Errorf("want %s,\n got %s", want, got)
	}
}

func TestConfigDefaults(t *testing.T) {
	ctx := context.TestContext(t)
	got := dir.Config(ctx)
	if got == "" {
		t.Errorf("want non-empty config dir")
	}
	if !strings.Contains(got, "kn") {
		t.Errorf("want config dir to contain 'kn'")
	}
}

func TestCache(t *testing.T) {
	ctx := context.TestContext(t)
	want := t.TempDir()
	ctx = dir.WithCacheDir(ctx, want)

	got := dir.Cache(ctx)
	if got != want {
		t.Errorf("want %s,\n got %s", want, got)
	}
}

func TestCacheDefaults(t *testing.T) {
	ctx := context.TestContext(t)
	got := dir.Cache(ctx)
	if got == "" {
		t.Errorf("want non-empty cache dir")
	}
	if !strings.Contains(got, "kn") {
		t.Errorf("want cache dir to contain 'kn'")
	}
}
