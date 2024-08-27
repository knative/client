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

package dir

import (
	"context"
	"testing"
)

// WithConfigDir returns a new context with the given config directory
// configured. It should be used for testing only.
func WithConfigDir(ctx context.Context, p string) context.Context {
	if !testing.Testing() {
		panic("WithConfigDir should be used only in tests")
	}
	return context.WithValue(ctx, configDirKey, p)
}

// WithCacheDir returns a new context with the given cache directory
// configured. It should be used for testing only.
func WithCacheDir(ctx context.Context, p string) context.Context {
	if !testing.Testing() {
		panic("WithCacheDir should be used only in tests")
	}
	return context.WithValue(ctx, cacheDirKey, p)
}
