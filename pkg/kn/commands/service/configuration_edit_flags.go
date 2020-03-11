// Copyright © 2018 The Knative Authors
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

package service

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"knative.dev/client/pkg/kn/flags"
	servinglib "knative.dev/client/pkg/serving"
	"knative.dev/client/pkg/util"
	"knative.dev/serving/pkg/apis/serving"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

type ConfigurationEditFlags struct {
	// Direct field manipulation
	Image   uniqueStringArg
	Env     []string
	EnvFrom []string
	Mount   []string
	Volume  []string

	Command string
	Arg     []string

	RequestsFlags, LimitsFlags ResourceFlags
	MinScale                   int
	MaxScale                   int
	ConcurrencyTarget          int
	ConcurrencyLimit           int
	AutoscaleWindow            string
	Port                       int32
	Labels                     []string
	LabelsService              []string
	LabelsRevision             []string
	NamePrefix                 string
	RevisionName               string
	ServiceAccountName         string
	ImagePullSecrets           string
	Annotations                []string
	ClusterLocal               bool
	User                       int64

	// Preferences about how to do the action.
	LockToDigest         bool
	GenerateRevisionName bool
	ForceCreate          bool

	// Bookkeeping
	flags []string
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

// markFlagMakesRevision indicates that a flag will create a new revision if you
// set it.
func (p *ConfigurationEditFlags) markFlagMakesRevision(f string) {
	p.flags = append(p.flags, f)
}

// addSharedFlags adds the flags common between create & update.
func (p *ConfigurationEditFlags) addSharedFlags(command *cobra.Command) {
	command.Flags().VarP(&p.Image, "image", "", "Image to run.")
	p.markFlagMakesRevision("image")
	command.Flags().StringArrayVarP(&p.Env, "env", "e", []string{},
		"Environment variable to set. NAME=value; you may provide this flag "+
			"any number of times to set multiple environment variables. "+
			"To unset, specify the environment variable name followed by a \"-\" (e.g., NAME-).")
	p.markFlagMakesRevision("env")

	command.Flags().StringArrayVarP(&p.EnvFrom, "env-from", "", []string{},
		"Add environment variables from a ConfigMap (prefix cm: or config-map:) or a Secret (prefix secret:). "+
			"Example: --env-from cm:myconfigmap or --env-from secret:mysecret. "+
			"You can use this flag multiple times. "+
			"To unset a ConfigMap/Secret reference, append \"-\" to the name, e.g. --env-from cm:myconfigmap-.")
	p.markFlagMakesRevision("env-from")

	command.Flags().StringArrayVarP(&p.Mount, "mount", "", []string{},
		"Mount a ConfigMap (prefix cm: or config-map:), a Secret (prefix secret: or sc:), or an existing Volume (without any prefix) on the specified directory. "+
			"Example: --mount /mydir=cm:myconfigmap, --mount /mydir=secret:mysecret, or --mount /mydir=myvolume. "+
			"When a configmap or a secret is specified, a corresponding volume is automatically generated. "+
			"You can use this flag multiple times. "+
			"For unmounting a directory, append \"-\", e.g. --mount /mydir-, which also removes any auto-generated volume.")
	p.markFlagMakesRevision("mount")

	command.Flags().StringArrayVarP(&p.Volume, "volume", "", []string{},
		"Add a volume from a ConfigMap (prefix cm: or config-map:) or a Secret (prefix secret: or sc:). "+
			"Example: --volume myvolume=cm:myconfigmap or --volume myvolume=secret:mysecret. "+
			"You can use this flag multiple times. "+
			"To unset a ConfigMap/Secret reference, append \"-\" to the name, e.g. --volume myvolume-.")
	p.markFlagMakesRevision("volume")

	command.Flags().StringVarP(&p.Command, "cmd", "", "",
		"Specify command to be used as entrypoint instead of default one. "+
			"Example: --cmd /app/start or --cmd /app/start --arg myArg to pass aditional arguments.")
	p.markFlagMakesRevision("cmd")

	command.Flags().StringArrayVarP(&p.Arg, "arg", "", []string{},
		"Add argument to the container command. "+
			"Example: --arg myArg1 --arg --myArg2 --arg myArg3=3. "+
			"You can use this flag multiple times.")
	p.markFlagMakesRevision("arg")

	command.Flags().StringVar(&p.RequestsFlags.CPU, "requests-cpu", "", "The requested CPU (e.g., 250m).")
	p.markFlagMakesRevision("requests-cpu")

	command.Flags().StringVar(&p.RequestsFlags.Memory, "requests-memory", "", "The requested memory (e.g., 64Mi).")
	p.markFlagMakesRevision("requests-memory")

	command.Flags().StringVar(&p.LimitsFlags.CPU, "limits-cpu", "", "The limits on the requested CPU (e.g., 1000m).")
	p.markFlagMakesRevision("limits-cpu")

	command.Flags().StringVar(&p.LimitsFlags.Memory, "limits-memory", "",
		"The limits on the requested memory (e.g., 1024Mi).")
	p.markFlagMakesRevision("limits-memory")

	command.Flags().IntVar(&p.MinScale, "min-scale", 0, "Minimal number of replicas.")
	p.markFlagMakesRevision("min-scale")

	command.Flags().IntVar(&p.MaxScale, "max-scale", 0, "Maximal number of replicas.")
	p.markFlagMakesRevision("max-scale")

	command.Flags().StringVar(&p.AutoscaleWindow, "autoscale-window", "", "Duration to look back for making auto-scaling decisions. The service is scaled to zero if no request was received in during that time. (eg: 10s)")
	p.markFlagMakesRevision("autoscale-window")

	flags.AddBothBoolFlagsUnhidden(command.Flags(), &p.ClusterLocal, "cluster-local", "", false,
		"Specify that the service be private. (--no-cluster-local will make the service publicly available)")
	//TODO: Need to also not change revision when already set (solution to issue #646)
	p.markFlagMakesRevision("cluster-local")
	p.markFlagMakesRevision("no-cluster-local")

	command.Flags().IntVar(&p.ConcurrencyTarget, "concurrency-target", 0,
		"Recommendation for when to scale up based on the concurrent number of incoming request. "+
			"Defaults to --concurrency-limit when given.")
	p.markFlagMakesRevision("concurrency-target")

	command.Flags().IntVar(&p.ConcurrencyLimit, "concurrency-limit", 0,
		"Hard Limit of concurrent requests to be processed by a single replica.")
	p.markFlagMakesRevision("concurrency-limit")

	command.Flags().Int32VarP(&p.Port, "port", "p", 0, "The port where application listens on.")
	p.markFlagMakesRevision("port")

	command.Flags().StringArrayVarP(&p.Labels, "label", "l", []string{},
		"Labels to set for both Service and Revision. name=value; you may provide this flag "+
			"any number of times to set multiple labels. "+
			"To unset, specify the label name followed by a \"-\" (e.g., name-).")
	p.markFlagMakesRevision("label")

	command.Flags().StringArrayVarP(&p.LabelsService, "label-service", "", []string{},
		"Service label to set. name=value; you may provide this flag "+
			"any number of times to set multiple labels. "+
			"To unset, specify the label name followed by a \"-\" (e.g., name-). This flag takes "+
			"precedence over \"label\" flag.")
	p.markFlagMakesRevision("label-service")
	command.Flags().StringArrayVarP(&p.LabelsRevision, "label-revision", "", []string{},
		"Revision label to set. name=value; you may provide this flag "+
			"any number of times to set multiple labels. "+
			"To unset, specify the label name followed by a \"-\" (e.g., name-). This flag takes "+
			"precedence over \"label\" flag.")
	p.markFlagMakesRevision("label-revision")

	command.Flags().StringVar(&p.RevisionName, "revision-name", "{{.Service}}-{{.Random 5}}-{{.Generation}}",
		"The revision name to set. Must start with the service name and a dash as a prefix. "+
			"Empty revision name will result in the server generating a name for the revision. "+
			"Accepts golang templates, allowing {{.Service}} for the service name, "+
			"{{.Generation}} for the generation, and {{.Random [n]}} for n random consonants.")
	p.markFlagMakesRevision("revision-name")

	flags.AddBothBoolFlagsUnhidden(command.Flags(), &p.LockToDigest, "lock-to-digest", "", true,
		"Keep the running image for the service constant when not explicitly specifying "+
			"the image. (--no-lock-to-digest pulls the image tag afresh with each new revision)")
	// Don't mark as changing the revision.

	command.Flags().StringVar(&p.ServiceAccountName,
		"service-account",
		"",
		"Service account name to set. An empty argument (\"\") clears the service account. The referenced service account must exist in the service's namespace.")
	p.markFlagMakesRevision("service-account")

	command.Flags().StringArrayVar(&p.Annotations, "annotation", []string{},
		"Service annotation to set. name=value; you may provide this flag "+
			"any number of times to set multiple annotations. "+
			"To unset, specify the annotation name followed by a \"-\" (e.g., name-).")
	p.markFlagMakesRevision("annotation")

	command.Flags().StringVar(&p.ImagePullSecrets,
		"pull-secret",
		"",
		"Image pull secret to set. An empty argument (\"\") clears the pull secret. The referenced secret must exist in the service's namespace.")
	p.markFlagMakesRevision("pull-secret")
	command.Flags().Int64VarP(&p.User, "user", "", 0, "The user ID to run the container (e.g., 1001).")
	p.markFlagMakesRevision("user")
}

// AddUpdateFlags adds the flags specific to update.
func (p *ConfigurationEditFlags) AddUpdateFlags(command *cobra.Command) {
	p.addSharedFlags(command)
}

// AddCreateFlags adds the flags specific to create
func (p *ConfigurationEditFlags) AddCreateFlags(command *cobra.Command) {
	p.addSharedFlags(command)
	command.Flags().BoolVar(&p.ForceCreate, "force", false,
		"Create service forcefully, replaces existing service if any.")
	command.MarkFlagRequired("image")
}

// Apply mutates the given service according to the flags in the command.
func (p *ConfigurationEditFlags) Apply(
	service *servingv1.Service,
	baseRevision *servingv1.Revision,
	cmd *cobra.Command) error {

	template := &service.Spec.Template
	if cmd.Flags().Changed("env") {
		envMap, err := util.MapFromArrayAllowingSingles(p.Env, "=")
		if err != nil {
			return errors.Wrap(err, "Invalid --env")
		}

		envToRemove := util.ParseMinusSuffix(envMap)
		err = servinglib.UpdateEnvVars(template, envMap, envToRemove)
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

		err := servinglib.UpdateEnvFrom(template, envFromSourceToUpdate, envFromSourceToRemove)
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("mount") || cmd.Flags().Changed("volume") {
		mountsToUpdate, mountsToRemove, err := util.OrderedMapAndRemovalListFromArray(p.Mount, "=")
		if err != nil {
			return errors.Wrap(err, "Invalid --mount")
		}

		volumesToUpdate, volumesToRemove, err := util.OrderedMapAndRemovalListFromArray(p.Volume, "=")
		if err != nil {
			return errors.Wrap(err, "Invalid --volume")
		}

		err = servinglib.UpdateVolumeMountsAndVolumes(template, mountsToUpdate, mountsToRemove, volumesToUpdate, volumesToRemove)
		if err != nil {
			return err
		}
	}

	name, err := servinglib.GenerateRevisionName(p.RevisionName, service)
	if err != nil {
		return err
	}

	if p.AnyMutation(cmd) {
		template.Name = name
	}

	imageSet := false
	if cmd.Flags().Changed("image") {
		err = servinglib.UpdateImage(template, p.Image.String())
		if err != nil {
			return err
		}
		imageSet = true
	}
	_, userImagePresent := template.Annotations[servinglib.UserImageAnnotationKey]
	freezeMode := userImagePresent || cmd.Flags().Changed("lock-to-digest")
	if p.LockToDigest && p.AnyMutation(cmd) && freezeMode {
		servinglib.SetUserImageAnnot(template)
		if !imageSet {
			err = servinglib.FreezeImageToDigest(template, baseRevision)
			if err != nil {
				return err
			}
		}
	} else if !p.LockToDigest {
		servinglib.UnsetUserImageAnnot(template)
	}

	limitsResources, err := p.computeResources(p.LimitsFlags)
	if err != nil {
		return err
	}
	requestsResources, err := p.computeResources(p.RequestsFlags)
	if err != nil {
		return err
	}
	err = servinglib.UpdateResources(template, requestsResources, limitsResources)
	if err != nil {
		return err
	}

	if cmd.Flags().Changed("cmd") {
		err = servinglib.UpdateContainerCommand(template, p.Command)
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("arg") {
		err = servinglib.UpdateContainerArg(template, p.Arg)
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("port") {
		err = servinglib.UpdateContainerPort(template, p.Port)
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("min-scale") {
		err = servinglib.UpdateMinScale(template, p.MinScale)
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("max-scale") {
		err = servinglib.UpdateMaxScale(template, p.MaxScale)
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("autoscale-window") {
		err = servinglib.UpdateAutoscaleWindow(template, p.AutoscaleWindow)
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("concurrency-target") {
		err = servinglib.UpdateConcurrencyTarget(template, p.ConcurrencyTarget)
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("concurrency-limit") {
		err = servinglib.UpdateConcurrencyLimit(template, int64(p.ConcurrencyLimit))
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("cluster-local") || cmd.Flags().Changed("no-cluster-local") {
		if p.ClusterLocal {
			labels := servinglib.UpdateLabels(service.ObjectMeta.Labels, map[string]string{serving.VisibilityLabelKey: serving.VisibilityClusterLocal}, []string{})
			service.ObjectMeta.Labels = labels // In case service.ObjectMeta.Labels was nil
		} else {
			labels := servinglib.UpdateLabels(service.ObjectMeta.Labels, map[string]string{}, []string{serving.VisibilityLabelKey})
			service.ObjectMeta.Labels = labels // In case service.ObjectMeta.Labels was nil
		}
	}

	if cmd.Flags().Changed("label") || cmd.Flags().Changed("label-service") || cmd.Flags().Changed("label-revision") {
		labelsAllMap, err := util.MapFromArrayAllowingSingles(p.Labels, "=")
		if err != nil {
			return errors.Wrap(err, "Invalid --label")
		}

		err = p.updateLabels(&service.ObjectMeta, p.LabelsService, labelsAllMap)
		if err != nil {
			return errors.Wrap(err, "Invalid --label-service")
		}

		err = p.updateLabels(&template.ObjectMeta, p.LabelsRevision, labelsAllMap)
		if err != nil {
			return errors.Wrap(err, "Invalid --label-revision")
		}
	}

	if cmd.Flags().Changed("annotation") {
		annotationsMap, err := util.MapFromArrayAllowingSingles(p.Annotations, "=")
		if err != nil {
			return errors.Wrap(err, "Invalid --annotation")
		}

		annotationsToRemove := util.ParseMinusSuffix(annotationsMap)
		err = servinglib.UpdateAnnotations(service, template, annotationsMap, annotationsToRemove)
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("service-account") {
		err = servinglib.UpdateServiceAccountName(template, p.ServiceAccountName)
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("pull-secret") {
		servinglib.UpdateImagePullSecrets(template, p.ImagePullSecrets)
	}

	if cmd.Flags().Changed("user") {
		servinglib.UpdateUser(template, p.User)
	}

	return nil
}

func (p *ConfigurationEditFlags) updateLabels(obj *metav1.ObjectMeta, flagLabels []string, labelsAllMap map[string]string) error {
	labelFlagMap, err := util.MapFromArrayAllowingSingles(flagLabels, "=")
	if err != nil {
		return errors.Wrap(err, "Unable to parse label flags")
	}
	labelsMap := make(util.StringMap)
	labelsMap.Merge(labelsAllMap)
	labelsMap.Merge(labelFlagMap)
	revisionLabelsToRemove := util.ParseMinusSuffix(labelsMap)
	obj.Labels = servinglib.UpdateLabels(obj.Labels, labelsMap, revisionLabelsToRemove)

	return nil
}

func (p *ConfigurationEditFlags) computeResources(resourceFlags ResourceFlags) (corev1.ResourceList, error) {
	resourceList := corev1.ResourceList{}

	if resourceFlags.CPU != "" {
		cpuQuantity, err := resource.ParseQuantity(resourceFlags.CPU)
		if err != nil {
			return corev1.ResourceList{},
				errors.Wrapf(err, "Error parsing %q", resourceFlags.CPU)
		}

		resourceList[corev1.ResourceCPU] = cpuQuantity
	}

	if resourceFlags.Memory != "" {
		memoryQuantity, err := resource.ParseQuantity(resourceFlags.Memory)
		if err != nil {
			return corev1.ResourceList{},
				errors.Wrapf(err, "Error parsing %q", resourceFlags.Memory)
		}

		resourceList[corev1.ResourceMemory] = memoryQuantity
	}

	return resourceList, nil
}

// AnyMutation returns true if there are any revision template mutations in the
// command.
func (p *ConfigurationEditFlags) AnyMutation(cmd *cobra.Command) bool {
	for _, flag := range p.flags {
		if cmd.Flags().Changed(flag) {
			return true
		}
	}
	return false
}
