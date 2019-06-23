/*
Copyright 2019 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e

import (
	"flag"
)

// ClientFlags saves the environment flags for client.
var ClientFlags = initializeFlags()

type ClientEnvironmentFlags struct {
	EmitMetrics bool // Emit metrics
}

func initializeFlags() *ClientEnvironmentFlags {
	var f ClientEnvironmentFlags

	// emitmetrics is a required flag for running periodic test jobs, add it here to avoid the error
	flag.BoolVar(&f.EmitMetrics, "emitmetrics", false,
		"Set this flag to true if you would like tests to emit metrics, e.g. latency of resources being realized in the system.")
	flag.Parse()

	return &f
}
