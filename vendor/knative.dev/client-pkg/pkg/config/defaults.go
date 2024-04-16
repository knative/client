// Copyright © 2021 The Knative Authors
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

package config

import (
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

var (
	// DefaultRetry is the default set of rules
	// to be followed when retrying for conflicts during
	// resource update. Values are the same as
	// k8s.io/client-go/util/retry/util.go
	// except where commented on
	DefaultRetry = wait.Backoff{
		// Start retries with 20ms instead of 10ms
		Duration: 20 * time.Millisecond,
		// Increase the retry duration by 50%
		Factor: 1.5,
		Jitter: 0.1,
		Steps:  5,
	}
)
