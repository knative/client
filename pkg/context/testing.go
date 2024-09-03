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

package context

import (
	sysctx "context"

	"go.uber.org/zap/zaptest"
)

type testingTKey struct{}

// TestingT is a subset of the API provided by all *testing.T and *testing.B
// objects. It is also compatible with zaptest.TestingT. This allows us to
// write tests that use zaptest and our own context package.
type TestingT interface {
	Log(args ...any)
	Cleanup(func())
	TempDir() string
	Setenv(key, value string)

	zaptest.TestingT
}

// TestContext returns a new context with the given testing object configured.
func TestContext(t TestingT) sysctx.Context {
	return WithTestingT(sysctx.TODO(), t)
}

// WithTestingT returns a context with the given testing object configured.
func WithTestingT(ctx sysctx.Context, t TestingT) sysctx.Context {
	return sysctx.WithValue(ctx, testingTKey{}, t)
}

// TestingTFromContext returns the testing object configured in the given
// context. If no testing object is configured, it returns nil.
func TestingTFromContext(ctx sysctx.Context) TestingT {
	t, ok := ctx.Value(testingTKey{}).(TestingT)
	if !ok {
		return nil
	}
	return t
}
