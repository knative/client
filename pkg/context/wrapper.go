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
	"time"
)

type Context = sysctx.Context

// TODO is a wrapper for sysctx.TODO for ease of use (single import).
func TODO() sysctx.Context {
	return sysctx.TODO()
}

// Background is a wrapper for sysctx.Background for ease of use (single import).
func Background() sysctx.Context {
	return sysctx.Background()
}

// WithCancel is a wrapper for sysctx.WithCancel for ease of use (single import).
func WithCancel(parent sysctx.Context) (sysctx.Context, sysctx.CancelFunc) {
	return sysctx.WithCancel(parent)
}

// WithDeadline is a wrapper for sysctx.WithDeadline for ease of use (single import).
func WithDeadline(parent sysctx.Context, deadline time.Time) (sysctx.Context, sysctx.CancelFunc) {
	return sysctx.WithDeadline(parent, deadline)
}

// WithTimeout is a wrapper for sysctx.WithTimeout for ease of use (single import).
func WithTimeout(parent sysctx.Context, timeout time.Duration) (sysctx.Context, sysctx.CancelFunc) {
	return sysctx.WithTimeout(parent, timeout)
}

// WithValue is a wrapper for sysctx.WithValue for ease of use (single import).
func WithValue(parent sysctx.Context, key, val interface{}) sysctx.Context {
	return sysctx.WithValue(parent, key, val)
}

// Value is a wrapper for sysctx.Value for ease of use (single import).
func Value(ctx sysctx.Context, key interface{}) interface{} {
	return ctx.Value(key)
}
