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
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	knflags "knative.dev/client/pkg/kn/flags"
	servinglib "knative.dev/client/pkg/serving"
	"knative.dev/client/pkg/util"
	network "knative.dev/networking/pkg"
	"knative.dev/serving/pkg/apis/autoscaling"
	"knative.dev/serving/pkg/apis/serving"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

type ConfigurationEditFlags struct {
	//Fields for PodSpecFlags
	PodSpecFlags knflags.PodSpecFlags

	// Direct field manipulation
	Scale                  string
	MinScale               int
	MaxScale               int
	ConcurrencyTarget      int
	ConcurrencyLimit       int
	ConcurrencyUtilization int
	AutoscaleWindow        string
	Labels                 []string
	LabelsService          []string
	LabelsRevision         []string
	RevisionName           string
	Annotations            []string
	AnnotationsService     []string
	AnnotationsRevision    []string
	ClusterLocal           bool
	ScaleInit              int

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

	command.Flags().IntVar(&p.MinScale, "min-scale", 0, "Minimal number of replicas.")
	command.Flags().MarkHidden("min-scale")
	p.markFlagMakesRevision("min-scale")

	command.Flags().IntVar(&p.MaxScale, "max-scale", 0, "Maximal number of replicas.")
	command.Flags().MarkHidden("max-scale")
	p.markFlagMakesRevision("max-scale")

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

	command.Flags().StringVar(&p.AutoscaleWindow, "autoscale-window", "", "Duration to look back for making auto-scaling decisions. The service is scaled to zero if no request was received in during that time. (eg: 10s)")
	p.markFlagMakesRevision("autoscale-window")

	knflags.AddBothBoolFlagsUnhidden(command.Flags(), &p.ClusterLocal, "cluster-local", "", false,
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

	command.Flags().IntVar(&p.ConcurrencyUtilization, "concurrency-utilization", 70,
		"Percentage of concurrent requests utilization before scaling up.")
	p.markFlagMakesRevision("concurrency-utilization")

	command.Flags().StringArrayVarP(&p.Labels, "label", "l", []string{},
		"Labels to set for both Service and Revision. name=value; you may provide this flag "+
			"any number of times to set multiple labels. "+
			"To unset, specify the label name followed by a \"-\" (e.g., name-).")
	p.markFlagMakesRevision("label")

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

	command.Flags().StringArrayVarP(&p.Annotations, "annotation", "a", []string{},
		"Annotations to set for both Service and Revision. name=value; you may provide this flag "+
			"any number of times to set multiple annotations. "+
			"To unset, specify the annotation name followed by a \"-\" (e.g., name-).")
	p.markFlagMakesRevision("annotation")

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
	command.Flags().StringVarP(&p.Filename, "filename", "f", "", "Create a service from file. "+
		"The created service can be further modified by combining with other options. "+
		"For example, -f /path/to/file --env NAME=value adds also an environment variable.")
	command.MarkFlagFilename("filename")
	p.markFlagMakesRevision("filename")
}

// Apply mutates the given service according to the flags in the command.
func (p *ConfigurationEditFlags) Apply(
	service *servingv1.Service,
	baseRevision *servingv1.Revision,
	cmd *cobra.Command) error {

	template := &service.Spec.Template

	err := p.PodSpecFlags.ResolvePodSpec(&template.Spec.PodSpec, cmd.Flags())
	if err != nil {
		return err
	}

	name := ""
	if p.RevisionName != "" {
		name, err = servinglib.GenerateRevisionName(p.RevisionName, service)
		if err != nil {
			return err
		}

		if p.AnyMutation(cmd) {
			template.Name = name
		}
	}

	_, userImagePresent := template.Annotations[servinglib.UserImageAnnotationKey]
	freezeMode := userImagePresent || cmd.Flags().Changed("lock-to-digest")
	if p.LockToDigest && p.AnyMutation(cmd) && freezeMode {
		servinglib.SetUserImageAnnot(template)
		if !cmd.Flags().Changed("image") {
			err = servinglib.FreezeImageToDigest(template, baseRevision)
			if err != nil {
				return err
			}
		}
	}

	if !p.LockToDigest {
		servinglib.UnsetUserImageAnnot(template)
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

	if cmd.Flags().Changed("concurrency-utilization") {
		err = servinglib.UpdateConcurrencyUtilization(template, p.ConcurrencyUtilization)
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

	return nil
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
