/*
Copyright 2025 The Knative Authors

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

// This file contains logic to encapsulate flags which are needed to specify
// what cluster, etc. to use for e2e tests.

package e2e

import (
	"flag"

	// Load the generic flags of knative.dev/pkg too.
	_ "knative.dev/pkg/test"
)

// ClientFlags holds the flags or defaults for knative/client settings in the user's environment.
var ClientFlags = initializeClientFlags()

// ClientEnvironmentFlags holds the e2e flags needed only by the client repo.
type ClientEnvironmentFlags struct {
	HTTPS bool // Indicates where the test service will be created with https
}

func initializeClientFlags() *ClientEnvironmentFlags {
	var f ClientEnvironmentFlags

	flag.BoolVar(&f.HTTPS, "https", false,
		"Set this flag to true to run all tests with https.")

	return &f
}
