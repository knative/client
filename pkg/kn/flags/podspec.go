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
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"knative.dev/client/pkg/util"

	"github.com/spf13/pflag"
)

// PodSpecFlags to hold the container resource requirements values
type PodSpecFlags struct {
	// Direct field manipulation
	Image           uniqueStringArg
	ImagePullPolicy string
	Env             []string
	EnvFrom         []string
	EnvValueFrom    []string
	EnvFile         string
	Mount           []string
	Volume          []string

	Command []string
	Arg     []string

	ExtraContainers string

	Resources          ResourceOptions
	Port               string
	ServiceAccountName string
	ImagePullSecrets   string
	User               int64
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

//AddUpdateFlags will add PodSpec flags related to environment variable to FlagSet of update command
func (p *PodSpecFlags) AddUpdateFlags(flagset *pflag.FlagSet) []string {
	flagNames := []string{}
	flagset.StringArrayVarP(&p.Env, "env", "e", []string{},
		"Environment variable to set. NAME=value; you may provide this flag "+
			"any number of times to set multiple environment variables. "+
			"To unset, specify the environment variable name followed by a \"-\" (e.g., NAME-).")
	flagNames = append(flagNames, "env")

	flagset.StringArrayVarP(&p.EnvValueFrom, "env-value-from", "", []string{},
		"Add environment variable from a value of key in ConfigMap (prefix cm: or config-map:) or a Secret (prefix sc: or secret:). "+
			"Example: --env-value-from NAME=cm:myconfigmap:key or --env-value-from NAME=secret:mysecret:key. "+
			"You can use this flag multiple times. "+
			"To unset a value from a ConfigMap/Secret key reference, append \"-\" to the key, e.g. --env-value-from ENV-.")
	flagNames = append(flagNames, "env-value-from")

	flagset.StringArrayVarP(&p.EnvFrom, "env-from", "", []string{},
		"Add environment variables from a ConfigMap (prefix cm: or config-map:) or a Secret (prefix secret:). "+
			"Example: --env-from cm:myconfigmap or --env-from secret:mysecret. "+
			"You can use this flag multiple times. "+
			"To unset a ConfigMap/Secret reference, append \"-\" to the name, e.g. --env-from cm:myconfigmap-.")
	flagNames = append(flagNames, "env-from")

	return flagNames
}

//AddCreateFlags will add PodSpec flags related to environment variable to FlagSet of create command
func (p *PodSpecFlags) AddCreateFlags(flagset *pflag.FlagSet) []string {
	flagNames := []string{}
	flagset.StringArrayVarP(&p.Env, "env", "e", []string{},
		"Environment variable to set. NAME=value; you may provide this flag "+
			"any number of times to set multiple environment variables.")
	flagNames = append(flagNames, "env")

	flagset.StringArrayVarP(&p.EnvValueFrom, "env-value-from", "", []string{},
		"Add environment variable from a value of key in ConfigMap (prefix cm: or config-map:) or a Secret (prefix sc: or secret:). "+
			"Example: --env-value-from NAME=cm:myconfigmap:key or --env-value-from NAME=secret:mysecret:key. "+
			"You can use this flag multiple times.")
	flagNames = append(flagNames, "env-value-from")

	flagset.StringArrayVarP(&p.EnvFrom, "env-from", "", []string{},
		"Add environment variables from a ConfigMap (prefix cm: or config-map:) or a Secret (prefix secret:). "+
			"Example: --env-from cm:myconfigmap or --env-from secret:mysecret. "+
			"You can use this flag multiple times.")
	flagNames = append(flagNames, "env-from")

	return flagNames
}

//AddFlags will add PodSpec related flags to FlagSet
func (p *PodSpecFlags) AddFlags(flagset *pflag.FlagSet) []string {

	flagNames := []string{}

	flagset.VarP(&p.Image, "image", "", "Image to run.")
	flagNames = append(flagNames, "image")

	flagset.StringVar(&p.ImagePullPolicy, "pull-policy", "",
		"Image pull policy. Valid values (case insensitive): Always | Never | IfNotPresent")

	flagset.StringVarP(&p.EnvFile, "env-file", "", "", "Path to a file containing environment variables (e.g. --env-file=/home/knative/service1/env).")
	flagNames = append(flagNames, "env-file")

	flagset.StringArrayVarP(&p.Mount, "mount", "", []string{},
		"Mount a ConfigMap (prefix cm: or config-map:), a Secret (prefix secret: or sc:), an EmptyDir (prefix ed: or emptyDir:)"+
			"or an existing Volume (without any prefix) on the specified directory. "+
			"Example: --mount /mydir=cm:myconfigmap, --mount /mydir=secret:mysecret, --mount /mydir=emptyDir:myvol "+
			"or --mount /mydir=myvolume. When a configmap or a secret is specified, a corresponding volume is "+
			"automatically generated. You can specify a volume subpath by following the volume name with slash separated path. "+
			"Example: --mount /mydir=cm:myconfigmap/subpath/to/be/mounted. "+
			"You can use this flag multiple times. "+
			"For unmounting a directory, append \"-\", e.g. --mount /mydir-, which also removes any auto-generated volume.")
	flagNames = append(flagNames, "mount")

	flagset.StringArrayVarP(&p.Volume, "volume", "", []string{},
		"Add a volume from a ConfigMap (prefix cm: or config-map:) a Secret (prefix secret: or sc:), or "+
			"an EmptyDir (prefix ed: or emptyDir:). "+
			"Example: --volume myvolume=cm:myconfigmap, --volume myvolume=secret:mysecret or --volume emptyDir:myvol:size=1Gi,type=Memory. "+
			"You can use this flag multiple times. "+
			"To unset a ConfigMap/Secret reference, append \"-\" to the name, e.g. --volume myvolume-.")
	flagNames = append(flagNames, "volume")

	flagset.StringArrayVarP(&p.Command, "cmd", "", []string{},
		"Specify command to be used as entrypoint instead of default one. "+
			"Example: --cmd /app/start or --cmd sh --cmd /app/start.sh or --cmd /app/start --arg myArg to pass additional arguments.")
	flagNames = append(flagNames, "cmd")

	flagset.StringArrayVarP(&p.Arg, "arg", "", []string{},
		"Add argument to the container command. "+
			"Example: --arg myArg1 --arg --myArg2 --arg myArg3=3. "+
			"You can use this flag multiple times.")
	flagNames = append(flagNames, "arg")

	// DEPRECATED since 1.0
	flagset.StringVarP(&p.ExtraContainers, "extra-containers", "", "",
		"Deprecated, use --containers instead.")
	flagset.MarkHidden("extra-containers")
	flagNames = append(flagNames, "containers")

	flagset.StringVarP(&p.ExtraContainers, "containers", "", "",
		"Specify path to file including definition for additional containers, alternatively use '-' to read from stdin. "+
			"Example: --containers ./containers.yaml or --containers -.")
	flagNames = append(flagNames, "containers")

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

// ResolvePodSpec will create corev1.PodSpec based on the flag inputs and all input arguments
func (p *PodSpecFlags) ResolvePodSpec(podSpec *corev1.PodSpec, flags *pflag.FlagSet, allArgs []string) error {
	var err error

	if flags.Changed("env") || flags.Changed("env-value-from") || flags.Changed("env-file") {
		envToUpdate, envToRemove, err := util.OrderedMapAndRemovalListFromArray(p.Env, "=")
		if err != nil {
			return fmt.Errorf("Invalid --env: %w", err)
		}

		envValueFromToUpdate, envValueFromToRemove, err := util.OrderedMapAndRemovalListFromArray(p.EnvValueFrom, "=")
		if err != nil {
			return fmt.Errorf("Invalid --env-value-from: %w", err)
		}

		envsFileToUpdate := util.NewOrderedMap()
		envsFileToRemove := []string{}
		if p.EnvFile != "" {
			envsFromFile, err := util.GetEnvsFromFile(p.EnvFile, "=")
			if err != nil {
				return fmt.Errorf("Invalid --env-file: %w", err)
			}
			envsFileToUpdate, envsFileToRemove, err = util.OrderedMapAndRemovalListFromArray(envsFromFile, "=")
			if err != nil {
				return fmt.Errorf("Invalid --env: %w", err)
			}
		}

		err = UpdateEnvVars(
			podSpec, allArgs, envToUpdate, envToRemove,
			envValueFromToUpdate, envValueFromToRemove,
			p.EnvFile, envsFileToUpdate, envsFileToRemove,
		)
		if err != nil {
			return err
		}
	}

	if flags.Changed("env-from") {
		envFromSourceToUpdate := []string{}
		envFromSourceToRemove := []string{}
		for _, name := range p.EnvFrom {
			if name == "-" {
				return fmt.Errorf("\"-\" is not a valid value for \"--env-from\"")
			} else if strings.HasSuffix(name, "-") {
				envFromSourceToRemove = append(envFromSourceToRemove, name[:len(name)-1])
			} else {
				envFromSourceToUpdate = append(envFromSourceToUpdate, name)
			}
		}

		err := UpdateEnvFrom(podSpec, envFromSourceToUpdate, envFromSourceToRemove)
		if err != nil {
			return err
		}
	}

	if flags.Changed("mount") || flags.Changed("volume") {
		mountsToUpdate, mountsToRemove, err := util.OrderedMapAndRemovalListFromArray(p.Mount, "=")
		if err != nil {
			return fmt.Errorf("Invalid --mount: %w", err)
		}

		volumesToUpdate, volumesToRemove, err := util.OrderedMapAndRemovalListFromArray(p.Volume, "=")
		if err != nil {
			return fmt.Errorf("Invalid --volume: %w", err)
		}

		err = UpdateVolumeMountsAndVolumes(podSpec, mountsToUpdate, mountsToRemove, volumesToUpdate, volumesToRemove)
		if err != nil {
			return err
		}
	}

	if flags.Changed("image") {
		err = UpdateImage(podSpec, p.Image.String())
		if err != nil {
			return err
		}
	}

	if flags.Changed("pull-policy") {

		err = UpdateImagePullPolicy(podSpec, p.ImagePullPolicy)
		if err != nil {
			return err
		}
	}

	requestsToRemove, limitsToRemove, err := p.Resources.Validate()
	if err != nil {
		return err
	}

	err = UpdateResources(podSpec, p.Resources.ResourceRequirements, requestsToRemove, limitsToRemove)
	if err != nil {
		return err
	}

	if flags.Changed("cmd") {
		err = UpdateContainerCommand(podSpec, p.Command)
		if err != nil {
			return err
		}
	}

	if flags.Changed("arg") {
		err = UpdateContainerArg(podSpec, p.Arg)
		if err != nil {
			return err
		}
	}

	if flags.Changed("port") {
		err = UpdateContainerPort(podSpec, p.Port)
		if err != nil {
			return err
		}
	}

	if flags.Changed("service-account") {
		UpdateServiceAccountName(podSpec, p.ServiceAccountName)
	}

	if flags.Changed("pull-secret") {
		UpdateImagePullSecrets(podSpec, p.ImagePullSecrets)
	}

	if flags.Changed("user") {
		err = UpdateUser(podSpec, p.User)
		if err != nil {
			return err
		}
	}

	if flags.Changed("containers") || flags.Changed("extra-containers") || p.ExtraContainers == "-" {
		var fromFile *corev1.PodSpec
		fromFile, err = decodeContainersFromFile(p.ExtraContainers)
		if err != nil {
			return err
		}
		UpdateContainers(podSpec, fromFile.Containers)
	}

	return nil
}
