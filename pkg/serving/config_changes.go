// Copyright © 2019 The Knative Authors
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

package serving

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"knative.dev/pkg/ptr"
	"knative.dev/serving/pkg/apis/autoscaling"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"knative.dev/client/pkg/util"
)

// VolumeSourceType is a type standing for enumeration of ConfigMap and Secret
type VolumeSourceType int

// Enumeration of volume source types: ConfigMap or Secret
const (
	ConfigMapVolumeSourceType VolumeSourceType = iota
	SecretVolumeSourceType
	PortFormatErr = "the port specification '%s' is not valid. Please provide in the format 'NAME:PORT', where 'NAME' is optional. Examples: '--port h2c:8080' , '--port 8080'."
)

var (
	UserImageAnnotationKey = "client.knative.dev/user-image"
	ApiTooOldError         = errors.New("the service is using too old of an API format for the operation")
)

func (vt VolumeSourceType) String() string {
	names := [...]string{"config-map", "secret"}
	if vt < ConfigMapVolumeSourceType || vt > SecretVolumeSourceType {
		return "unknown"
	}
	return names[vt]
}

// UpdateEnvVars gives the configuration all the env var values listed in the given map of
// vars.  Does not touch any environment variables not mentioned, but it can add
// new env vars and change the values of existing ones, then sort by env key name.
func UpdateEnvVars(template *servingv1.RevisionTemplateSpec, toUpdate map[string]string, toRemove []string) error {
	container, err := ContainerOfRevisionTemplate(template)
	if err != nil {
		return err
	}
	updated := updateEnvVarsFromMap(container.Env, toUpdate)
	updated = removeEnvVars(updated, toRemove)
	// Sort by env key name
	sort.SliceStable(updated, func(i, j int) bool {
		return updated[i].Name < updated[j].Name
	})
	container.Env = updated

	return nil
}

// UpdateEnvFrom updates envFrom
func UpdateEnvFrom(template *servingv1.RevisionTemplateSpec, toUpdate []string, toRemove []string) error {
	container, err := ContainerOfRevisionTemplate(template)
	if err != nil {
		return err
	}
	envFrom, err := updateEnvFrom(container.EnvFrom, toUpdate)
	if err != nil {
		return err
	}
	container.EnvFrom, err = removeEnvFrom(envFrom, toRemove)
	return err
}

func reviseVolumeInfoAndMountsToUpdate(volumes []corev1.Volume, mountsToUpdate *util.OrderedMap,
	volumesToUpdate *util.OrderedMap) (*util.OrderedMap, *util.OrderedMap, error) {
	volumeSourceInfoByName := util.NewOrderedMap() //make(map[string]*volumeSourceInfo)
	mountsToUpdateRevised := util.NewOrderedMap()  //make(map[string]string)

	it := mountsToUpdate.Iterator()
	for path, value, ok := it.NextString(); ok; path, value, ok = it.NextString() {
		// slices[0] -> config-map, cm, secret, sc, volume, or vo
		// slices[1] -> secret, config-map, or volume name
		slices := strings.SplitN(value, ":", 2)
		if len(slices) == 1 {
			mountsToUpdateRevised.Set(path, slices[0])
		} else {
			switch volumeType := slices[0]; volumeType {
			case "config-map", "cm":
				generatedName := GenerateVolumeName(path)
				volumeSourceInfoByName.Set(generatedName, &volumeSourceInfo{
					volumeSourceType: ConfigMapVolumeSourceType,
					volumeSourceName: slices[1],
				})
				mountsToUpdateRevised.Set(path, generatedName)
			case "secret", "sc":
				generatedName := GenerateVolumeName(path)
				volumeSourceInfoByName.Set(generatedName, &volumeSourceInfo{
					volumeSourceType: SecretVolumeSourceType,
					volumeSourceName: slices[1],
				})
				mountsToUpdateRevised.Set(path, generatedName)

			default:
				return nil, nil, fmt.Errorf("unsupported volume type \"%q\"; supported volume types are \"config-map or cm\", \"secret or sc\", and \"volume or vo\"", slices[0])
			}
		}
	}

	it = volumesToUpdate.Iterator()
	for name, value, ok := it.NextString(); ok; name, value, ok = it.NextString() {
		info, err := newVolumeSourceInfoWithSpecString(value)
		if err != nil {
			return nil, nil, err
		}
		volumeSourceInfoByName.Set(name, info)
	}

	return volumeSourceInfoByName, mountsToUpdateRevised, nil
}

func reviseVolumesToRemove(volumeMounts []corev1.VolumeMount, volumesToRemove []string, mountsToRemove []string) []string {
	for _, pathToRemove := range mountsToRemove {
		for _, volumeMount := range volumeMounts {
			if volumeMount.MountPath == pathToRemove && volumeMount.Name == GenerateVolumeName(pathToRemove) {
				volumesToRemove = append(volumesToRemove, volumeMount.Name)
			}
		}
	}
	return volumesToRemove
}

// UpdateVolumeMountsAndVolumes updates the configuration for volume mounts and volumes.
func UpdateVolumeMountsAndVolumes(template *servingv1.RevisionTemplateSpec,
	mountsToUpdate *util.OrderedMap, mountsToRemove []string, volumesToUpdate *util.OrderedMap, volumesToRemove []string) error {
	container, err := ContainerOfRevisionTemplate(template)
	if err != nil {
		return err
	}

	volumeSourceInfoByName, mountsToUpdate, err := reviseVolumeInfoAndMountsToUpdate(template.Spec.Volumes, mountsToUpdate, volumesToUpdate)
	if err != nil {
		return err
	}

	volumes, err := updateVolumesFromMap(template.Spec.Volumes, volumeSourceInfoByName)
	if err != nil {
		return err
	}

	volumeMounts, err := updateVolumeMountsFromMap(container.VolumeMounts, mountsToUpdate, volumes)
	if err != nil {
		return err
	}

	volumesToRemove = reviseVolumesToRemove(container.VolumeMounts, volumesToRemove, mountsToRemove)

	container.VolumeMounts = removeVolumeMounts(volumeMounts, mountsToRemove)
	template.Spec.Volumes, err = removeVolumes(volumes, volumesToRemove, container.VolumeMounts)

	return err
}

// UpdateMinScale updates min scale annotation
func UpdateMinScale(template *servingv1.RevisionTemplateSpec, min int) error {
	return UpdateRevisionTemplateAnnotation(template, autoscaling.MinScaleAnnotationKey, strconv.Itoa(min))
}

// UpdateMaxScale updates max scale annotation
func UpdateMaxScale(template *servingv1.RevisionTemplateSpec, max int) error {
	return UpdateRevisionTemplateAnnotation(template, autoscaling.MaxScaleAnnotationKey, strconv.Itoa(max))
}

// UpdateAutoscaleWindow updates the autoscale window annotation
func UpdateAutoscaleWindow(template *servingv1.RevisionTemplateSpec, window string) error {
	_, err := time.ParseDuration(window)
	if err != nil {
		return fmt.Errorf("invalid duration for 'autoscale-window': %v", err)
	}
	return UpdateRevisionTemplateAnnotation(template, autoscaling.WindowAnnotationKey, window)
}

// UpdateConcurrencyTarget updates container concurrency annotation
func UpdateConcurrencyTarget(template *servingv1.RevisionTemplateSpec, target int) error {
	return UpdateRevisionTemplateAnnotation(template, autoscaling.TargetAnnotationKey, strconv.Itoa(target))
}

// UpdateConcurrencyUtilization updates container target utilization percentage annotation
func UpdateConcurrencyUtilization(template *servingv1.RevisionTemplateSpec, target int) error {
	return UpdateRevisionTemplateAnnotation(template, autoscaling.TargetUtilizationPercentageKey, strconv.Itoa(target))
}

// UpdateConcurrencyLimit updates container concurrency limit
func UpdateConcurrencyLimit(template *servingv1.RevisionTemplateSpec, limit int64) error {
	if limit < 0 {
		return fmt.Errorf("invalid concurrency-limit %d (must not be less than 0)", limit)
	}
	template.Spec.ContainerConcurrency = ptr.Int64(limit)
	return nil
}

// UpdateRevisionTemplateAnnotation updates an annotation for the given Revision Template.
// Also validates the autoscaling annotation values
func UpdateRevisionTemplateAnnotation(template *servingv1.RevisionTemplateSpec, annotation string, value string) error {
	annoMap := template.Annotations
	if annoMap == nil {
		annoMap = make(map[string]string)
		template.Annotations = annoMap
	}

	// Validate autoscaling annotations and returns error if invalid input provided
	// without changing the existing spec
	in := make(map[string]string)
	in[annotation] = value
	// The boolean indicates whether or not the init-scale annotation can be set to 0.
	// Since we don't have the config handy, err towards allowing it. The API will
	// correctly fail the request if it's forbidden.
	if err := autoscaling.ValidateAnnotations(true, in); err != nil {
		return err
	}

	annoMap[annotation] = value
	return nil
}

// EnvToMap is an utility function to translate between the API list form of env vars, and the
// more convenient map form.
func EnvToMap(vars []corev1.EnvVar) (map[string]string, error) {
	result := map[string]string{}
	for _, envVar := range vars {
		_, present := result[envVar.Name]
		if present {
			return nil, fmt.Errorf("env var name present more than once: %v", envVar.Name)
		}
		result[envVar.Name] = envVar.Value
	}
	return result, nil
}

// UpdateImage a given image
func UpdateImage(template *servingv1.RevisionTemplateSpec, image string) error {
	// When not setting the image to a digest, add the user image annotation.
	container, err := ContainerOfRevisionTemplate(template)
	if err != nil {
		return err
	}
	container.Image = image
	return nil
}

// UnsetUserImageAnnot removes the user image annotation
func UnsetUserImageAnnot(template *servingv1.RevisionTemplateSpec) {
	delete(template.Annotations, UserImageAnnotationKey)
}

// SetUserImageAnnot sets the user image annotation if the image isn't by-digest already.
func SetUserImageAnnot(template *servingv1.RevisionTemplateSpec) {
	// If the current image isn't by-digest, set the user-image annotation to it
	// so we remember what it was.
	currentContainer, _ := ContainerOfRevisionTemplate(template)
	ui := currentContainer.Image
	if strings.Contains(ui, "@") {
		prev, ok := template.Annotations[UserImageAnnotationKey]
		if ok {
			ui = prev
		}
	}
	if template.Annotations == nil {
		template.Annotations = make(map[string]string)
	}
	template.Annotations[UserImageAnnotationKey] = ui
}

// FreezeImageToDigest sets the image on the template to the image digest of the base revision.
func FreezeImageToDigest(template *servingv1.RevisionTemplateSpec, baseRevision *servingv1.Revision) error {
	if baseRevision == nil {
		return nil
	}

	currentContainer, err := ContainerOfRevisionTemplate(template)
	if err != nil {
		return err
	}

	baseContainer, err := ContainerOfRevisionSpec(&baseRevision.Spec)
	if err != nil {
		return err
	}
	if currentContainer.Image != baseContainer.Image {
		return fmt.Errorf("could not freeze image to digest since current revision contains unexpected image")
	}

	if baseRevision.Status.DeprecatedImageDigest != "" {
		return UpdateImage(template, baseRevision.Status.DeprecatedImageDigest)
	}
	return nil
}

// UpdateContainerCommand updates container with a given argument
func UpdateContainerCommand(template *servingv1.RevisionTemplateSpec, command string) error {
	container, err := ContainerOfRevisionTemplate(template)
	if err != nil {
		return err
	}
	container.Command = []string{command}
	return nil
}

// UpdateContainerArg updates container with a given argument
func UpdateContainerArg(template *servingv1.RevisionTemplateSpec, arg []string) error {
	container, err := ContainerOfRevisionTemplate(template)
	if err != nil {
		return err
	}
	container.Args = arg
	return nil
}

// UpdateContainerPort updates container with a given name:port
func UpdateContainerPort(template *servingv1.RevisionTemplateSpec, port string) error {
	container, err := ContainerOfRevisionTemplate(template)
	if err != nil {
		return err
	}

	var containerPort int64
	var name string

	elements := strings.SplitN(port, ":", 2)
	if len(elements) == 2 {
		name = elements[0]
		containerPort, err = strconv.ParseInt(elements[1], 10, 32)
		if err != nil {
			return fmt.Errorf(PortFormatErr, port)
		}
	} else {
		name = ""
		containerPort, err = strconv.ParseInt(elements[0], 10, 32)
		if err != nil {
			return fmt.Errorf(PortFormatErr, port)
		}
	}

	container.Ports = []corev1.ContainerPort{{
		ContainerPort: int32(containerPort),
		Name:          name,
	}}
	return nil
}

// UpdateRunAsUser updates container with a given user id
func UpdateUser(template *servingv1.RevisionTemplateSpec, user int64) error {
	container, err := ContainerOfRevisionTemplate(template)
	if err != nil {
		return err
	}
	container.SecurityContext = &corev1.SecurityContext{
		RunAsUser: &user,
	}
	return nil
}

// UpdateResources updates container resources for given revision template
func UpdateResources(template *servingv1.RevisionTemplateSpec, resources corev1.ResourceRequirements, requestsToRemove, limitsToRemove []string) error {
	container, err := ContainerOfRevisionTemplate(template)
	if err != nil {
		return err
	}

	if container.Resources.Requests == nil {
		container.Resources.Requests = corev1.ResourceList{}
	}

	for k, v := range resources.Requests {
		container.Resources.Requests[k] = v
	}

	for _, reqToRemove := range requestsToRemove {
		delete(container.Resources.Requests, corev1.ResourceName(reqToRemove))
	}

	if container.Resources.Limits == nil {
		container.Resources.Limits = corev1.ResourceList{}
	}

	for k, v := range resources.Limits {
		container.Resources.Limits[k] = v
	}

	for _, limToRemove := range limitsToRemove {
		delete(container.Resources.Limits, corev1.ResourceName(limToRemove))
	}

	return nil
}

// UpdateResourcesDeprecated updates resources as requested
func UpdateResourcesDeprecated(template *servingv1.RevisionTemplateSpec, requestsResourceList corev1.ResourceList, limitsResourceList corev1.ResourceList) error {
	container, err := ContainerOfRevisionTemplate(template)
	if err != nil {
		return err
	}
	if container.Resources.Requests == nil {
		container.Resources.Requests = corev1.ResourceList{}
	}

	for k, v := range requestsResourceList {
		container.Resources.Requests[k] = v
	}

	if container.Resources.Limits == nil {
		container.Resources.Limits = corev1.ResourceList{}
	}

	for k, v := range limitsResourceList {
		container.Resources.Limits[k] = v
	}

	return nil
}

// UpdateLabels updates the labels by adding items from `add` then removing any items from `remove`
func UpdateLabels(labelsMap map[string]string, add map[string]string, remove []string) map[string]string {
	if labelsMap == nil {
		labelsMap = map[string]string{}
	}

	for key, value := range add {
		labelsMap[key] = value
	}
	for _, key := range remove {
		delete(labelsMap, key)
	}

	return labelsMap
}

// UpdateAnnotations updates the annotations identically on a service and template.
// Does not overwrite the entire Annotations field, only makes the requested updates.
func UpdateAnnotations(
	service *servingv1.Service,
	template *servingv1.RevisionTemplateSpec,
	toUpdate map[string]string,
	toRemove []string) error {

	if service.ObjectMeta.Annotations == nil {
		service.ObjectMeta.Annotations = make(map[string]string)
	}

	if template.ObjectMeta.Annotations == nil {
		template.ObjectMeta.Annotations = make(map[string]string)
	}

	for key, value := range toUpdate {
		service.ObjectMeta.Annotations[key] = value
		template.ObjectMeta.Annotations[key] = value
	}

	for _, key := range toRemove {
		delete(service.ObjectMeta.Annotations, key)
		delete(template.ObjectMeta.Annotations, key)
	}

	return nil
}

// UpdateServiceAccountName updates the service account name used for the corresponding knative service
func UpdateServiceAccountName(template *servingv1.RevisionTemplateSpec, serviceAccountName string) error {
	serviceAccountName = strings.TrimSpace(serviceAccountName)
	template.Spec.ServiceAccountName = serviceAccountName
	return nil
}

// UpdateImagePullSecrets updates the image pull secrets used for the corresponding knative service
func UpdateImagePullSecrets(template *servingv1.RevisionTemplateSpec, pullsecrets string) {
	pullsecrets = strings.TrimSpace(pullsecrets)
	if pullsecrets == "" {
		template.Spec.ImagePullSecrets = nil
	} else {
		template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{{
			Name: pullsecrets,
		}}
	}
}

// GenerateVolumeName generates a volume name with respect to a given path string.
// Current implementation basically sanitizes the path string by replacing "/" with "-"
// To reduce any chance of duplication, a checksum part generated from the path string is appended to the sanitized string.
// The volume name must follow the DNS label standard as defined in RFC 1123. This means the name must:
// - contain at most 63 characters
// - contain only lowercase alphanumeric characters or '-'
// - start with an alphanumeric character
// - end with an alphanumeric character
func GenerateVolumeName(path string) string {
	builder := &strings.Builder{}
	for idx, r := range path {
		switch {
		case unicode.IsLower(r) || unicode.IsDigit(r) || r == '-':
			builder.WriteRune(r)
		case unicode.IsUpper(r):
			builder.WriteRune(unicode.ToLower(r))
		case r == '/':
			if idx != 0 {
				builder.WriteRune('-')
			}
		default:
			builder.WriteRune('-')
		}
	}

	vname := appendCheckSum(builder.String(), path)

	// the name must start with an alphanumeric character
	if !unicode.IsLetter(rune(vname[0])) && !unicode.IsNumber(rune(vname[0])) {
		vname = fmt.Sprintf("k-%s", vname)
	}

	// contain at most 63 characters
	if len(vname) > 63 {
		// must end with an alphanumeric character
		vname = fmt.Sprintf("%s-n", vname[0:61])
	}

	return vname
}

// =======================================================================================

func updateEnvVarsFromMap(env []corev1.EnvVar, toUpdate map[string]string) []corev1.EnvVar {
	set := sets.NewString()
	for i := range env {
		envVar := &env[i]
		if val, ok := toUpdate[envVar.Name]; ok {
			envVar.Value = val
			set.Insert(envVar.Name)
		}
	}
	for name, val := range toUpdate {
		if !set.Has(name) {
			env = append(env, corev1.EnvVar{Name: name, Value: val})
		}
	}
	return env
}

func removeEnvVars(env []corev1.EnvVar, toRemove []string) []corev1.EnvVar {
	for _, name := range toRemove {
		for i, envVar := range env {
			if envVar.Name == name {
				env = append(env[:i], env[i+1:]...)
				break
			}
		}
	}
	return env
}

func updateEnvFrom(envFromSources []corev1.EnvFromSource, toUpdate []string) ([]corev1.EnvFromSource, error) {
	existingNameSet := make(map[string]bool)

	for _, envSrc := range envFromSources {
		if canonicalName, err := getCanonicalNameFromEnvFromSource(&envSrc); err == nil {
			existingNameSet[canonicalName] = true
		}
	}

	for _, s := range toUpdate {
		info, err := newVolumeSourceInfoWithSpecString(s)
		if err != nil {
			return nil, err
		}

		if _, ok := existingNameSet[info.getCanonicalName()]; !ok {
			envFromSources = append(envFromSources, *info.createEnvFromSource())
		}
	}

	return envFromSources, nil
}

func removeEnvFrom(envFromSources []corev1.EnvFromSource, toRemove []string) ([]corev1.EnvFromSource, error) {
	for _, name := range toRemove {
		info, err := newVolumeSourceInfoWithSpecString(name)
		if err != nil {
			return nil, err
		}
		for i, envSrc := range envFromSources {
			if (info.volumeSourceType == ConfigMapVolumeSourceType && envSrc.ConfigMapRef != nil && info.volumeSourceName == envSrc.ConfigMapRef.Name) ||
				(info.volumeSourceType == SecretVolumeSourceType && envSrc.SecretRef != nil && info.volumeSourceName == envSrc.SecretRef.Name) {
				envFromSources = append(envFromSources[:i], envFromSources[i+1:]...)
				break
			}
		}
	}

	if len(envFromSources) == 0 {
		envFromSources = nil
	}

	return envFromSources, nil
}

func updateVolume(volume *corev1.Volume, info *volumeSourceInfo) error {
	switch info.volumeSourceType {
	case ConfigMapVolumeSourceType:
		volume.Secret = nil
		volume.ConfigMap = &corev1.ConfigMapVolumeSource{LocalObjectReference: corev1.LocalObjectReference{Name: info.volumeSourceName}}
	case SecretVolumeSourceType:
		volume.ConfigMap = nil
		volume.Secret = &corev1.SecretVolumeSource{SecretName: info.volumeSourceName}
	default:
		return fmt.Errorf("Invalid VolumeSourceType")
	}
	return nil
}

// updateVolumeMountsFromMap updates or adds volume mounts. If a given name of a volume is not existing, it returns an error
func updateVolumeMountsFromMap(volumeMounts []corev1.VolumeMount, toUpdate *util.OrderedMap, volumes []corev1.Volume) ([]corev1.VolumeMount, error) {
	set := make(map[string]bool)

	for i := range volumeMounts {
		volumeMount := &volumeMounts[i]
		name, present := toUpdate.GetString(volumeMount.MountPath)

		if present {
			if !existsVolumeNameInVolumes(name, volumes) {
				return nil, fmt.Errorf("There is no volume matched with %q", name)
			}

			volumeMount.ReadOnly = true
			volumeMount.Name = name
			set[volumeMount.MountPath] = true
		}
	}

	it := toUpdate.Iterator()
	for mountPath, name, ok := it.NextString(); ok; mountPath, name, ok = it.NextString() {
		if !set[mountPath] {
			volumeMounts = append(volumeMounts, corev1.VolumeMount{
				Name:      name,
				ReadOnly:  true,
				MountPath: mountPath,
			})
		}
	}

	return volumeMounts, nil
}

func removeVolumeMounts(volumeMounts []corev1.VolumeMount, toRemove []string) []corev1.VolumeMount {
	for _, mountPath := range toRemove {
		for i, volumeMount := range volumeMounts {
			if volumeMount.MountPath == mountPath {
				volumeMounts = append(volumeMounts[:i], volumeMounts[i+1:]...)
				break
			}
		}
	}

	if len(volumeMounts) == 0 {
		return nil
	}

	return volumeMounts
}

// updateVolumesFromMap updates or adds volumes regardless whether the volume is used or not
func updateVolumesFromMap(volumes []corev1.Volume, toUpdate *util.OrderedMap) ([]corev1.Volume, error) {
	set := make(map[string]bool)

	for i := range volumes {
		volume := &volumes[i]
		info, present := toUpdate.Get(volume.Name)
		if present {
			err := updateVolume(volume, info.(*volumeSourceInfo))
			if err != nil {
				return nil, err
			}
			set[volume.Name] = true
		}
	}

	it := toUpdate.Iterator()
	for name, info, ok := it.Next(); ok; name, info, ok = it.Next() {
		if !set[name] {
			volumes = append(volumes, corev1.Volume{Name: name})
			updateVolume(&volumes[len(volumes)-1], info.(*volumeSourceInfo))
		}
	}

	return volumes, nil
}

// removeVolumes removes volumes. If there is a volume mount referencing the volume, it causes an error
func removeVolumes(volumes []corev1.Volume, toRemove []string, volumeMounts []corev1.VolumeMount) ([]corev1.Volume, error) {
	for _, name := range toRemove {
		for i, volume := range volumes {
			if volume.Name == name {
				if existsVolumeNameInVolumeMounts(name, volumeMounts) {
					return nil, fmt.Errorf("The volume %q cannot be removed because it is mounted", name)
				}
				volumes = append(volumes[:i], volumes[i+1:]...)
				break
			}
		}
	}

	if len(volumes) == 0 {
		return nil, nil
	}

	return volumes, nil
}

// =======================================================================================

type volumeSourceInfo struct {
	volumeSourceType VolumeSourceType
	volumeSourceName string
}

func newVolumeSourceInfoWithSpecString(spec string) (*volumeSourceInfo, error) {
	slices := strings.SplitN(spec, ":", 2)
	if len(slices) != 2 {
		return nil, fmt.Errorf("argument requires a value that contains the : character; got %q", spec)
	}

	var volumeSourceType VolumeSourceType

	typeString := strings.TrimSpace(slices[0])
	volumeSourceName := strings.TrimSpace(slices[1])

	switch typeString {
	case "config-map", "cm":
		volumeSourceType = ConfigMapVolumeSourceType
	case "secret", "sc":
		volumeSourceType = SecretVolumeSourceType
	default:
		return nil, fmt.Errorf("unsupported volume source type \"%q\"; supported volume source types are \"config-map\" and \"secret\"", slices[0])
	}

	if len(volumeSourceName) == 0 {
		return nil, fmt.Errorf("the name of %s cannot be an empty string", volumeSourceType)
	}

	return &volumeSourceInfo{
		volumeSourceType: volumeSourceType,
		volumeSourceName: volumeSourceName,
	}, nil
}

func (vol *volumeSourceInfo) getCanonicalName() string {
	return fmt.Sprintf("%s:%s", vol.volumeSourceType, vol.volumeSourceName)
}

func getCanonicalNameFromEnvFromSource(envSrc *corev1.EnvFromSource) (string, error) {
	if envSrc.ConfigMapRef != nil {
		return fmt.Sprintf("%s:%s", ConfigMapVolumeSourceType, envSrc.ConfigMapRef.Name), nil
	}
	if envSrc.SecretRef != nil {
		return fmt.Sprintf("%s:%s", SecretVolumeSourceType, envSrc.SecretRef.Name), nil
	}

	return "", fmt.Errorf("there is no ConfigMapRef or SecretRef in a EnvFromSource")
}

func (vol *volumeSourceInfo) createEnvFromSource() *corev1.EnvFromSource {
	switch vol.volumeSourceType {
	case ConfigMapVolumeSourceType:
		return &corev1.EnvFromSource{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: vol.volumeSourceName,
				}}}
	case SecretVolumeSourceType:
		return &corev1.EnvFromSource{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: vol.volumeSourceName,
				}}}
	}

	return nil
}

// =======================================================================================

func existsVolumeNameInVolumes(volumeName string, volumes []corev1.Volume) bool {
	for _, volume := range volumes {
		if volume.Name == volumeName {
			return true
		}
	}
	return false
}

func existsVolumeNameInVolumeMounts(volumeName string, volumeMounts []corev1.VolumeMount) bool {
	for _, volumeMount := range volumeMounts {
		if volumeMount.Name == volumeName {
			return true
		}
	}
	return false
}

func appendCheckSum(sanitizedString, path string) string {
	checkSum := sha1.Sum([]byte(path))
	shortCheckSum := checkSum[0:4]
	return fmt.Sprintf("%s-%x", sanitizedString, shortCheckSum)
}
