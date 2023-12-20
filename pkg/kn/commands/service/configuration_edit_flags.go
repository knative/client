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
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/serving/pkg/apis/config"

	knconfig "knative.dev/client/pkg/kn/config"
	knflags "knative.dev/client/pkg/kn/flags"
	servinglib "knative.dev/client/pkg/serving"
	"knative.dev/client/pkg/util"
	network "knative.dev/networking/pkg/apis/networking"
	"knative.dev/serving/pkg/apis/autoscaling"
	"knative.dev/serving/pkg/apis/serving"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

type ConfigurationEditFlags struct {
	//Fields for PodSpecFlags
	PodSpecFlags knflags.PodSpecFlags

	// Direct field manipulation
	Scale               string
	MinScale            int
	MaxScale            int
	ScaleActivation     int
	ScaleTarget         int
	ScaleMetric         string
	ConcurrencyLimit    int
	ScaleUtilization    int
	ScaleWindow         string
	Labels              []string
	LabelsService       []string
	LabelsRevision      []string
	RevisionName        string
	Annotations         []string
	AnnotationsService  []string
	AnnotationsRevision []string
	ClusterLocal        bool
	ScaleInit           int
	TimeoutSeconds      int64
	Profile             string

	// Preferences about how to do the action.
	LockToDigest         bool
	GenerateRevisionName bool
	ForceCreate          bool

	Filename string

	// Bookkeeping
	flags []string
}

// markFlagMakesRevision indicates that a flag will create a new revision if you
// set it.
func (p *ConfigurationEditFlags) markFlagMakesRevision(f string) {
	p.flags = append(p.flags, f)
}

// addSharedFlags adds the flags common between create & update.
func (p *ConfigurationEditFlags) addSharedFlags(command *cobra.Command) {
	flagNames := p.PodSpecFlags.AddFlags(command.Flags())
	for _, name := range flagNames {
		p.markFlagMakesRevision(name)
	}

	command.Flags().StringVar(&p.Scale, "scale", "",
		"Set the Minimum and Maximum number of replicas. You can use this flag to set both to a single value, "+
			"or set a range with min/max values, or set either min or max values without specifying the other. "+
			"Example: --scale 5 (scale-min = 5, scale-max = 5) or --scale 1..5 (scale-min = 1, scale-max = 5) or --scale "+
			"1.. (scale-min = 1, scale-max = unchanged) or --scale ..5 (scale-min = unchanged, scale-max = 5)")
	p.markFlagMakesRevision("scale")

	command.Flags().IntVar(&p.MinScale, "scale-min", 0, "Minimum number of replicas.")
	p.markFlagMakesRevision("scale-min")

	command.Flags().IntVar(&p.MaxScale, "scale-max", 0, "Maximum number of replicas.")
	p.markFlagMakesRevision("scale-max")

	command.Flags().IntVar(&p.ScaleActivation, "scale-activation", 0, "Minimum non-zero value that a service should scale to.")
	p.markFlagMakesRevision("scale-activation")

	command.Flags().StringVar(&p.ScaleMetric, "scale-metric", "", "Set the name of the metric the PodAutoscaler should scale on. "+
		"Example: --scale-metric rps (to scale on rps) or --scale-metric concurrency (to scale on concurrency). The default metric is concurrency.")
	p.markFlagMakesRevision("scale-metric")

	// DEPRECATED since 1.0
	command.Flags().StringVar(&p.ScaleWindow, "autoscale-window", "", "Deprecated option, please use --scale-window")
	p.markFlagMakesRevision("autoscale-window")
	command.Flags().MarkHidden("autoscale-window")

	command.Flags().StringVar(&p.ScaleWindow, "scale-window", "", "Duration to look back for making auto-scaling decisions. The service is scaled to zero if no request was received in during that time. (eg: 10s)")
	p.markFlagMakesRevision("scale-window")

	knflags.AddBothBoolFlagsUnhidden(command.Flags(), &p.ClusterLocal, "cluster-local", "", false,
		"Specify that the service be private. (--no-cluster-local will make the service publicly available)")

	// DEPRECATED since 1.0
	command.Flags().IntVar(&p.ScaleTarget, "concurrency-target", 0,
		"Deprecated, use --scale-target instead.")
	p.markFlagMakesRevision("concurrency-target")
	command.Flags().MarkHidden("concurrency-target")

	command.Flags().IntVar(&p.ScaleTarget, "scale-target", 0,
		"Recommendation for what metric value the PodAutoscaler should attempt to maintain. "+
			"Use with --scale-metric flag to configure the metric name for which the target value should be maintained. "+
			"Default metric name is concurrency. "+
			"The flag defaults to --concurrency-limit when given.")
	p.markFlagMakesRevision("scale-target")

	command.Flags().IntVar(&p.ConcurrencyLimit, "concurrency-limit", 0,
		"Hard Limit of concurrent requests to be processed by a single replica.")
	p.markFlagMakesRevision("concurrency-limit")

	// DEPRECATED since 1.0
	command.Flags().IntVar(&p.ScaleUtilization, "concurrency-utilization", 70,
		"Deprecated, use --scale-utilization instead.")
	p.markFlagMakesRevision("concurrency-utilization")
	command.Flags().MarkHidden("concurrency-utilization")

	command.Flags().IntVar(&p.ScaleUtilization, "scale-utilization", 70,
		"Percentage of concurrent requests utilization before scaling up.")
	p.markFlagMakesRevision("scale-utilization")

	command.Flags().StringArrayVarP(&p.LabelsService, "label-service", "", []string{},
		"Service label to set. name=value; you may provide this flag "+
			"any number of times to set multiple labels. "+
			"To unset, specify the label name followed by a \"-\" (e.g., name-). This flag takes "+
			"precedence over the \"label\" flag.")
	p.markFlagMakesRevision("label-service")
	command.Flags().StringArrayVarP(&p.LabelsRevision, "label-revision", "", []string{},
		"Revision label to set. name=value; you may provide this flag "+
			"any number of times to set multiple labels. "+
			"To unset, specify the label name followed by a \"-\" (e.g., name-). This flag takes "+
			"precedence over the \"label\" flag.")
	p.markFlagMakesRevision("label-revision")

	command.Flags().StringVar(&p.RevisionName, "revision-name", "",
		"The revision name to set. Must start with the service name and a dash as a prefix. "+
			"Empty revision name will result in the server generating a name for the revision. "+
			"Accepts golang templates, allowing {{.Service}} for the service name, "+
			"{{.Generation}} for the generation, and {{.Random [n]}} for n random consonants "+
			"(e.g. {{.Service}}-{{.Random 5}}-{{.Generation}})")
	p.markFlagMakesRevision("revision-name")

	knflags.AddBothBoolFlagsUnhidden(command.Flags(), &p.LockToDigest, "lock-to-digest", "", true,
		"Keep the running image for the service constant when not explicitly specifying "+
			"the image. (--no-lock-to-digest pulls the image tag afresh with each new revision)")
	// Don't mark as changing the revision.

	command.Flags().StringArrayVarP(&p.AnnotationsService, "annotation-service", "", []string{},
		"Service annotation to set. name=value; you may provide this flag "+
			"any number of times to set multiple annotations. "+
			"To unset, specify the annotation name followed by a \"-\" (e.g., name-). This flag takes "+
			"precedence over the \"annotation\" flag.")
	p.markFlagMakesRevision("annotation-service")

	command.Flags().StringArrayVarP(&p.AnnotationsRevision, "annotation-revision", "", []string{},
		"Revision annotation to set. name=value; you may provide this flag "+
			"any number of times to set multiple annotations. "+
			"To unset, specify the annotation name followed by a \"-\" (e.g., name-). This flag takes "+
			"precedence over the \"annotation\" flag.")
	p.markFlagMakesRevision("annotation-revision")

	command.Flags().IntVar(&p.ScaleInit, "scale-init", 0, "Initial number of replicas with which a service starts. Can be 0 or a positive integer.")
	p.markFlagMakesRevision("scale-init")

	command.Flags().Int64Var(&p.TimeoutSeconds, "timeout", config.DefaultRevisionTimeoutSeconds,
		"Duration in seconds that the request routing layer will wait for a request delivered to a "+""+
			"container to begin replying")
	p.markFlagMakesRevision("timeout")

	command.Flags().StringVar(&p.Profile, "profile", "",
		"The profile name must be defined in config.yaml or part of the built-in profile, e.g. Istio. Related annotations and labels will be added to the service."+
			"To unset, specify the profile name followed by a \"-\" (e.g., name-).")
	p.markFlagMakesRevision("profile")
}

// AddUpdateFlags adds the flags specific to update.
func (p *ConfigurationEditFlags) AddUpdateFlags(command *cobra.Command) {
	p.addSharedFlags(command)

	flagNames := p.PodSpecFlags.AddUpdateFlags(command.Flags())
	for _, name := range flagNames {
		p.markFlagMakesRevision(name)
	}

	command.Flags().StringArrayVarP(&p.Annotations, "annotation", "a", []string{},
		"Annotations to set for both Service and Revision. name=value; you may provide this flag "+
			"any number of times to set multiple annotations. "+
			"To unset, specify the annotation name followed by a \"-\" (e.g., name-).")
	p.markFlagMakesRevision("annotation")
	command.Flags().StringArrayVarP(&p.Labels, "label", "l", []string{},
		"Labels to set for both Service and Revision. name=value; you may provide this flag "+
			"any number of times to set multiple labels. "+
			"To unset, specify the label name followed by a \"-\" (e.g., name-).")
	p.markFlagMakesRevision("label")
}

// AddCreateFlags adds the flags specific to create
func (p *ConfigurationEditFlags) AddCreateFlags(command *cobra.Command) {
	p.addSharedFlags(command)

	flagNames := p.PodSpecFlags.AddCreateFlags(command.Flags())
	for _, name := range flagNames {
		p.markFlagMakesRevision(name)
	}

	command.Flags().BoolVar(&p.ForceCreate, "force", false,
		"Create service forcefully, replaces existing service if any.")
	command.Flags().StringVarP(&p.Filename, "filename", "f", "", "Create a service from file. "+
		"The created service can be further modified by combining with other options. "+
		"For example, -f /path/to/file --env NAME=value adds also an environment variable.")
	command.MarkFlagFilename("filename")
	p.markFlagMakesRevision("filename")
	command.Flags().StringArrayVarP(&p.Annotations, "annotation", "a", []string{},
		"Annotations to set for both Service and Revision. name=value; you may provide this flag "+
			"any number of times to set multiple annotations.")
	p.markFlagMakesRevision("annotation")
	command.Flags().StringArrayVarP(&p.Labels, "label", "l", []string{},
		"Labels to set for both Service and Revision. name=value; you may provide this flag "+
			"any number of times to set multiple labels.")
	p.markFlagMakesRevision("label")
}

// Apply mutates the given service according to the flags in the command.
func (p *ConfigurationEditFlags) Apply(
	service *servingv1.Service,
	baseRevision *servingv1.Revision,
	cmd *cobra.Command) error {

	template := &service.Spec.Template

	err := p.PodSpecFlags.ResolvePodSpec(&template.Spec.PodSpec, cmd.Flags(), os.Args)
	if err != nil {
		return err
	}

	// Client side revision naming requires to set a new revisionname
	if p.RevisionName != "" {
		name, err := servinglib.GenerateRevisionName(p.RevisionName, service)
		if err != nil {
			return err
		}

		if p.AnyMutation(cmd) {
			template.Name = name
		}
	}

	// If some change happened that can cause a revision, set the update timestamp
	// But not for "apply", this would destroy idempotency
	if p.AnyMutation(cmd) && cmd.Name() != "apply" {
		servinglib.UpdateTimestampAnnotation(template)
	}

	if p.shouldPinToImageDigest(template, cmd) {
		servinglib.UpdateUserImageAnnotation(template)
		// Don't copy over digest of base revision if an image is specified.
		// If an --image is given, always use the tagged named to cause a re-resolving
		// of the digest by the serving backend (except when you use "apply" where you
		// always have to provide an image
		if !cmd.Flags().Changed("image") || cmd.Name() == "apply" {
			err = servinglib.PinImageToDigest(template, baseRevision)
			if err != nil {
				return err
			}
		}
	}

	if !p.LockToDigest {
		servinglib.UnsetUserImageAnnotation(template)
	}

	// Deprecated "min-scale" in 0.19, updated to "scale-min"
	if cmd.Flags().Changed("scale-min") || cmd.Flags().Changed("min-scale") {
		err = servinglib.UpdateMinScale(template, p.MinScale)
		if err != nil {
			return err
		}
	}

	// Deprecated "max-scale" in 0.19, updated to "scale-max"
	if cmd.Flags().Changed("scale-max") || cmd.Flags().Changed("max-scale") {
		err = servinglib.UpdateMaxScale(template, p.MaxScale)
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("scale") {
		if cmd.Flags().Changed("scale-max") {
			return fmt.Errorf("only --scale or --scale-max can be specified")
		} else if cmd.Flags().Changed("scale-min") {
			return fmt.Errorf("only --scale or --scale-min can be specified")
		}
		if strings.Contains(p.Scale, "..") {
			scaleParts := strings.Split(p.Scale, "..")
			if scaleParts[0] != "" {
				scaleMin, err := strconv.Atoi(scaleParts[0])
				if err != nil {
					return err
				}
				err = servinglib.UpdateMinScale(template, scaleMin)
				if err != nil {
					return err
				}
			}
			if scaleParts[1] != "" {
				scaleMax, err := strconv.Atoi(scaleParts[1])
				if err != nil {
					return err
				}
				err = servinglib.UpdateMaxScale(template, scaleMax)
				if err != nil {
					return err
				}
			}
		} else if scaleMin, err := strconv.Atoi(p.Scale); err == nil {
			scaleMax := scaleMin
			err = servinglib.UpdateMaxScale(template, scaleMax)
			if err != nil {
				return err
			}
			err = servinglib.UpdateMinScale(template, scaleMin)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("Scale must be of the format x..y or x")
		}
	}

	if cmd.Flags().Changed("scale-window") || cmd.Flags().Changed("autoscale-window") {
		err = servinglib.UpdateScaleWindow(template, p.ScaleWindow)
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("scale-target") || cmd.Flags().Changed("concurrency-target") {
		err = servinglib.UpdateScaleTarget(template, p.ScaleTarget)
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("scale-metric") {
		servinglib.UpdateScaleMetric(template, p.ScaleMetric)
	}

	if cmd.Flags().Changed("scale-activation") {
		servinglib.UpdateScaleActivation(template, p.ScaleActivation)
	}

	if cmd.Flags().Changed("concurrency-limit") {
		err = servinglib.UpdateConcurrencyLimit(template, int64(p.ConcurrencyLimit))
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("scale-utilization") || cmd.Flags().Changed("concurrency-utilization") {
		err = servinglib.UpdateScaleUtilization(template, p.ScaleUtilization)
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("cluster-local") || cmd.Flags().Changed("no-cluster-local") {
		if p.ClusterLocal {
			labels := servinglib.UpdateLabels(service.ObjectMeta.Labels, map[string]string{network.VisibilityLabelKey: serving.VisibilityClusterLocal}, []string{})
			service.ObjectMeta.Labels = labels // In case service.ObjectMeta.Labels was nil
		} else {
			labels := servinglib.UpdateLabels(service.ObjectMeta.Labels, map[string]string{}, []string{network.VisibilityLabelKey})
			service.ObjectMeta.Labels = labels // In case service.ObjectMeta.Labels was nil
		}
	}

	if cmd.Flags().Changed("label") || cmd.Flags().Changed("label-service") || cmd.Flags().Changed("label-revision") {
		labelsAllMap, err := util.MapFromArrayAllowingSingles(p.Labels, "=")
		if err != nil {
			return fmt.Errorf("Invalid --label: %w", err)
		}

		err = p.updateLabels(&service.ObjectMeta, p.LabelsService, labelsAllMap)
		if err != nil {
			return fmt.Errorf("Invalid --label-service: %w", err)
		}

		err = p.updateLabels(&template.ObjectMeta, p.LabelsRevision, labelsAllMap)
		if err != nil {
			return fmt.Errorf("Invalid --label-revision: %w", err)
		}
	}

	if cmd.Flags().Changed("annotation") || cmd.Flags().Changed("annotation-service") || cmd.Flags().Changed("annotation-revision") {
		annotationsAllMap, err := util.MapFromArrayAllowingSingles(p.Annotations, "=")
		if err != nil {
			return fmt.Errorf("Invalid --annotation: %w", err)
		}
		annotationRevisionFlagMap, err := util.MapFromArrayAllowingSingles(p.AnnotationsRevision, "=")
		if err != nil {
			return fmt.Errorf("Invalid --annotation-revision: %w", err)
		}
		annotationServiceFlagMap, err := util.MapFromArrayAllowingSingles(p.AnnotationsService, "=")
		if err != nil {
			return fmt.Errorf("Invalid --annotation-service: %w", err)
		}

		annotationsToRemove := util.ParseMinusSuffix(annotationsAllMap)

		revisionAnnotations := make(util.StringMap)
		revisionAnnotations.Merge(annotationsAllMap)
		revisionAnnotations.Merge(annotationRevisionFlagMap)
		err = servinglib.UpdateRevisionTemplateAnnotations(template, revisionAnnotations, annotationsToRemove)
		if err != nil {
			return err
		}

		serviceAnnotations := make(util.StringMap)

		// Service Annotations can't contain Autoscaling ones

		for key, value := range annotationsAllMap {
			if !strings.HasPrefix(key, autoscaling.GroupName) {
				serviceAnnotations[key] = value
			}
		}

		for key, value := range annotationServiceFlagMap {
			if strings.HasPrefix(key, autoscaling.GroupName) {
				return fmt.Errorf("service can not have auto-scaling related annotation: %s , please update using '--annotation-revision'", key)
			}
			serviceAnnotations[key] = value
		}

		err = servinglib.UpdateServiceAnnotations(service, serviceAnnotations, annotationsToRemove)
		if err != nil {
			return err
		}

	}

	if cmd.Flags().Changed("scale-init") {
		containsAnnotation := func(annotationList []string, annotation string) bool {
			for _, element := range annotationList {
				if strings.Contains(element, annotation) {
					return true
				}
			}
			return false
		}

		if cmd.Flags().Changed("annotation") && containsAnnotation(p.Annotations, autoscaling.InitialScaleAnnotationKey) {
			return fmt.Errorf("only one of the --scale-init or --annotation %s can be specified", autoscaling.InitialScaleAnnotationKey)
		}
		// Autoscaling annotations are only applicable on Revision Template, not Service
		err = servinglib.UpdateRevisionTemplateAnnotation(template, autoscaling.InitialScaleAnnotationKey, strconv.Itoa(p.ScaleInit))
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("timeout") {
		service.Spec.Template.Spec.TimeoutSeconds = &p.TimeoutSeconds
	}

	if cmd.Flags().Changed("profile") {
		profileName := ""
		deleteProfile := false
		if strings.HasSuffix(p.Profile, "-") {
			profileName = p.Profile[:len(p.Profile)-1]
			deleteProfile = true
		} else {
			profileName = p.Profile
			deleteProfile = false
		}

		if len(knconfig.GlobalConfig.Profile(profileName).Annotations) > 0 || len(knconfig.GlobalConfig.Profile(profileName).Labels) > 0 {
			annotations := knconfig.GlobalConfig.Profile(profileName).Annotations
			labels := knconfig.GlobalConfig.Profile(profileName).Labels

			profileAnnotations := make(util.StringMap)
			for _, value := range annotations {
				profileAnnotations[value.Name] = value.Value
			}

			profileLabels := make(util.StringMap)
			for _, value := range labels {
				profileLabels[value.Name] = value.Value
			}

			if deleteProfile {
				var annotationsToRemove []string
				for _, value := range annotations {
					annotationsToRemove = append(annotationsToRemove, value.Name)
				}

				var labelsToRemove []string
				for _, value := range labels {
					labelsToRemove = append(labelsToRemove, value.Name)
				}

				if err = servinglib.UpdateRevisionTemplateAnnotations(template, map[string]string{}, annotationsToRemove); err != nil {
					return err
				}
				updatedLabels := servinglib.UpdateLabels(service.ObjectMeta.Labels, map[string]string{}, labelsToRemove)
				service.ObjectMeta.Labels = updatedLabels // In case service.ObjectMeta.Labels was nil
			} else {
				if err = servinglib.UpdateRevisionTemplateAnnotations(template, profileAnnotations, []string{}); err != nil {
					return err
				}
				updatedLabels := servinglib.UpdateLabels(service.ObjectMeta.Labels, profileLabels, []string{})
				service.ObjectMeta.Labels = updatedLabels // In case service.ObjectMeta.Labels was nil
			}

			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("profile %s doesn't exist", profileName)
		}
	}

	return nil
}

// shouldPinToImageDigest checks whether the image digest that has been resolved in the current active
// revision should be used for the image update. This is useful if an update of the image
// should be prevented.
func (p *ConfigurationEditFlags) shouldPinToImageDigest(template *servingv1.RevisionTemplateSpec, cmd *cobra.Command) bool {
	// The user-image annotation is an indicator that the service has been
	// created with lock-to-digest. If this is not the case and neither --lock-to-digest
	// has been given explitly then we won't change the image
	_, userImagePresent := template.Annotations[servinglib.UserImageAnnotationKey]
	if !userImagePresent && !cmd.Flags().Changed("lock-to-digest") {
		return false
	}

	// When we want an update and --lock-to-digest is set (either given or the default), then
	// the image should be pinned to its digest
	return p.LockToDigest && p.AnyMutation(cmd)
}

func (p *ConfigurationEditFlags) updateLabels(obj *metav1.ObjectMeta, flagLabels []string, labelsAllMap map[string]string) error {
	labelFlagMap, err := util.MapFromArrayAllowingSingles(flagLabels, "=")
	if err != nil {
		return fmt.Errorf("Unable to parse label flags: %w", err)
	}
	labelsMap := make(util.StringMap)
	labelsMap.Merge(labelsAllMap)
	labelsMap.Merge(labelFlagMap)
	revisionLabelsToRemove := util.ParseMinusSuffix(labelsMap)
	obj.Labels = servinglib.UpdateLabels(obj.Labels, labelsMap, revisionLabelsToRemove)

	return nil
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
