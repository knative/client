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

package test

import (
	"flag"
	"os"
)

// Flags holds the command line flags or defaults for settings in the user's environment.
// See ClientFlags for the list of supported fields.
var Flags = InitializeFlags()

// ClientFlags define the flags that are needed to run the e2e tests.
type ClientFlags struct {
	DockerConfigJSON string
}

// InitializeFlags initializes the client's flags
func InitializeFlags() *ClientFlags {
	var f ClientFlags

	dockerConfigJSON := os.Getenv("DOCKER_CONFIG_JSON")
	flag.StringVar(&f.DockerConfigJSON, "dockerconfigjson", dockerConfigJSON,
		"Provide the path to Docker configuration file in json format. Defaults to $DOCKER_CONFIG_JSON")

	return &f
}
