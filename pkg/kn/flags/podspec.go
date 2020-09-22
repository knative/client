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
	"k8s.io/apimachinery/pkg/api/resource"
	"knative.dev/client/pkg/util"

	"github.com/spf13/cobra"
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
			"Example: --cmd /app/start or --cmd /app/start --arg myArg to pass additional arguments.")
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

func (p *PodSpecFlags) ResolvePodSpec(podSpec *corev1.PodSpec, cmd *cobra.Command) error {
	//var podSpec = &corev1.PodSpec{Containers: []corev1.Container{{}}}
	var err error

	if cmd.Flags().Changed("env") {
		envMap, err := util.MapFromArrayAllowingSingles(p.Env, "=")
		if err != nil {
			return fmt.Errorf("Invalid --env: %w", err)
		}

		envToRemove := util.ParseMinusSuffix(envMap)
		err = UpdateEnvVars(podSpec, envMap, envToRemove)
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("env-from") {
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

	if cmd.Flags().Changed("mount") || cmd.Flags().Changed("volume") {
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

	if cmd.Flags().Changed("image") {
		err = UpdateImage(podSpec, p.Image.String())
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("limits-cpu") || cmd.Flags().Changed("limits-memory") {
		if cmd.Flags().Changed("limit") {
			return fmt.Errorf("only one of (DEPRECATED) --limits-cpu / --limits-memory and --limit can be specified")
		}
		fmt.Fprintf(cmd.OutOrStdout(), "\nWARNING: flags --limits-cpu / --limits-memory are deprecated and going to be removed in future release, please use --limit instead.\n\n")
	}

	if cmd.Flags().Changed("requests-cpu") || cmd.Flags().Changed("requests-memory") {
		if cmd.Flags().Changed("request") {
			return fmt.Errorf("only one of (DEPRECATED) --requests-cpu / --requests-memory and --request can be specified")
		}
		fmt.Fprintf(cmd.OutOrStdout(), "\nWARNING: flags --requests-cpu / --requests-memory are deprecated and going to be removed in future release, please use --request instead.\n\n")
	}

	limitsResources, err := p.computeResources(p.LimitsFlags)
	if err != nil {
		return err
	}
	requestsResources, err := p.computeResources(p.RequestsFlags)
	if err != nil {
		return err
	}
	err = UpdateResourcesDeprecated(podSpec, requestsResources, limitsResources)
	if err != nil {
		return err
	}

	requestsToRemove, limitsToRemove, err := p.Resources.Validate()
	if err != nil {
		return err
	}

	err = UpdateResources(podSpec, p.Resources.ResourceRequirements, requestsToRemove, limitsToRemove)
	if err != nil {
		return err
	}

	if cmd.Flags().Changed("cmd") {
		err = UpdateContainerCommand(podSpec, p.Command)
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("arg") {
		err = UpdateContainerArg(podSpec, p.Arg)
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("port") {
		err = UpdateContainerPort(podSpec, p.Port)
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("service-account") {
		err = UpdateServiceAccountName(podSpec, p.ServiceAccountName)
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("pull-secret") {
		UpdateImagePullSecrets(podSpec, p.ImagePullSecrets)
	}

	if cmd.Flags().Changed("user") {
		UpdateUser(podSpec, p.User)
	}

	return nil
}

func (p *PodSpecFlags) computeResources(resourceFlags ResourceFlags) (corev1.ResourceList, error) {
	resourceList := corev1.ResourceList{}

	if resourceFlags.CPU != "" {
		cpuQuantity, err := resource.ParseQuantity(resourceFlags.CPU)
		if err != nil {
			return corev1.ResourceList{},
				fmt.Errorf("Error parsing %q: %w", resourceFlags.CPU, err)
		}

		resourceList[corev1.ResourceCPU] = cpuQuantity
	}

	if resourceFlags.Memory != "" {
		memoryQuantity, err := resource.ParseQuantity(resourceFlags.Memory)
		if err != nil {
			return corev1.ResourceList{},
				fmt.Errorf("Error parsing %q: %w", resourceFlags.Memory, err)
		}

		resourceList[corev1.ResourceMemory] = memoryQuantity
	}

	return resourceList, nil
}
