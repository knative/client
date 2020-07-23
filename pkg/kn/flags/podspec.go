// Copyright 2020 The Knative Authors
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

package flags

import (
	"errors"

	"github.com/spf13/pflag"
)

// PodSpecFlags to hold the container resource requirements values
type PodSpecFlags struct {
	// Direct field manipulation
	Image   uniqueStringArg
	Env     []string
	EnvFrom []string
	Mount   []string
	Volume  []string

	Command string
	Arg     []string

	RequestsFlags, LimitsFlags ResourceFlags // TODO: Flag marked deprecated in release v0.15.0, remove in release v0.18.0
	Resources                  ResourceOptions
	Port                       string
	ServiceAccountName         string
	ImagePullSecrets           string
	User                       int64
}

type ResourceFlags struct {
	CPU    string
	Memory string
}

// -- uniqueStringArg Value
// Custom implementation of flag.Value interface to prevent multiple value assignment.
// Useful to enforce unique use of flags, e.g. --image.
type uniqueStringArg string

func (s *uniqueStringArg) Set(val string) error {
	if len(*s) > 0 {
		return errors.New("can be provided only once")
	}
	*s = uniqueStringArg(val)
	return nil
}

func (s *uniqueStringArg) Type() string {
	return "string"
}

func (s *uniqueStringArg) String() string { return string(*s) }

//AddFlags will add PodSpec related flags to FlagSet
func (p *PodSpecFlags) AddFlags(flagset *pflag.FlagSet) []string {

	flagNames := []string{}

	flagset.VarP(&p.Image, "image", "", "Image to run.")
	flagNames = append(flagNames, "image")

	flagset.StringArrayVarP(&p.Env, "env", "e", []string{},
		"Environment variable to set. NAME=value; you may provide this flag "+
			"any number of times to set multiple environment variables. "+
			"To unset, specify the environment variable name followed by a \"-\" (e.g., NAME-).")
	flagNames = append(flagNames, "env")

	flagset.StringArrayVarP(&p.EnvFrom, "env-from", "", []string{},
		"Add environment variables from a ConfigMap (prefix cm: or config-map:) or a Secret (prefix secret:). "+
			"Example: --env-from cm:myconfigmap or --env-from secret:mysecret. "+
			"You can use this flag multiple times. "+
			"To unset a ConfigMap/Secret reference, append \"-\" to the name, e.g. --env-from cm:myconfigmap-.")
	flagNames = append(flagNames, "env-from")

	flagset.StringArrayVarP(&p.Mount, "mount", "", []string{},
		"Mount a ConfigMap (prefix cm: or config-map:), a Secret (prefix secret: or sc:), or an existing Volume (without any prefix) on the specified directory. "+
			"Example: --mount /mydir=cm:myconfigmap, --mount /mydir=secret:mysecret, or --mount /mydir=myvolume. "+
			"When a configmap or a secret is specified, a corresponding volume is automatically generated. "+
			"You can use this flag multiple times. "+
			"For unmounting a directory, append \"-\", e.g. --mount /mydir-, which also removes any auto-generated volume.")
	flagNames = append(flagNames, "mount")

	flagset.StringArrayVarP(&p.Volume, "volume", "", []string{},
		"Add a volume from a ConfigMap (prefix cm: or config-map:) or a Secret (prefix secret: or sc:). "+
			"Example: --volume myvolume=cm:myconfigmap or --volume myvolume=secret:mysecret. "+
			"You can use this flag multiple times. "+
			"To unset a ConfigMap/Secret reference, append \"-\" to the name, e.g. --volume myvolume-.")
	flagNames = append(flagNames, "volume")

	flagset.StringVarP(&p.Command, "cmd", "", "",
		"Specify command to be used as entrypoint instead of default one. "+
			"Example: --cmd /app/start or --cmd /app/start --arg myArg to pass aditional arguments.")
	flagNames = append(flagNames, "cmd")

	flagset.StringArrayVarP(&p.Arg, "arg", "", []string{},
		"Add argument to the container command. "+
			"Example: --arg myArg1 --arg --myArg2 --arg myArg3=3. "+
			"You can use this flag multiple times.")
	flagNames = append(flagNames, "arg")

	flagset.StringSliceVar(&p.Resources.Limits,
		"limit",
		nil,
		"The resource requirement limits for this Service. For example, 'cpu=100m,memory=256Mi'. "+
			"You can use this flag multiple times. "+
			"To unset a resource limit, append \"-\" to the resource name, e.g. '--limit memory-'.")
	flagNames = append(flagNames, "limit")

	flagset.StringSliceVar(&p.Resources.Requests,
		"request",
		nil,
		"The resource requirement requests for this Service. For example, 'cpu=100m,memory=256Mi'. "+
			"You can use this flag multiple times. "+
			"To unset a resource request, append \"-\" to the resource name, e.g. '--request cpu-'.")
	flagNames = append(flagNames, "request")

	flagset.StringVar(&p.RequestsFlags.CPU, "requests-cpu", "",
		"DEPRECATED: please use --request instead. The requested CPU (e.g., 250m).")
	flagNames = append(flagNames, "requests-cpu")

	flagset.StringVar(&p.RequestsFlags.Memory, "requests-memory", "",
		"DEPRECATED: please use --request instead. The requested memory (e.g., 64Mi).")
	flagNames = append(flagNames, "requests-memory")

	// TODO: Flag marked deprecated in release v0.15.0, remove in release v0.18.0
	flagset.StringVar(&p.LimitsFlags.CPU, "limits-cpu", "",
		"DEPRECATED: please use --limit instead. The limits on the requested CPU (e.g., 1000m).")
	flagNames = append(flagNames, "limits-cpu")

	// TODO: Flag marked deprecated in release v0.15.0, remove in release v0.18.0
	flagset.StringVar(&p.LimitsFlags.Memory, "limits-memory", "",
		"DEPRECATED: please use --limit instead. The limits on the requested memory (e.g., 1024Mi).")
	flagNames = append(flagNames, "limits-memory")

	flagset.StringVarP(&p.Port, "port", "p", "", "The port where application listens on, in the format 'NAME:PORT', where 'NAME' is optional. Examples: '--port h2c:8080' , '--port 8080'.")
	flagNames = append(flagNames, "port")

	flagset.StringVar(&p.ServiceAccountName,
		"service-account",
		"",
		"Service account name to set. An empty argument (\"\") clears the service account. The referenced service account must exist in the service's namespace.")
	flagNames = append(flagNames, "service-account")

	flagset.StringVar(&p.ImagePullSecrets,
		"pull-secret",
		"",
		"Image pull secret to set. An empty argument (\"\") clears the pull secret. The referenced secret must exist in the service's namespace.")
	flagNames = append(flagNames, "pull-secret")
	flagset.Int64VarP(&p.User, "user", "", 0, "The user ID to run the container (e.g., 1001).")
	flagNames = append(flagNames, "user")
	return flagNames
}
