// Copyright © 2020 The Knative Authors
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
	"fmt"
	"os"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/yaml"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"knative.dev/client/pkg/util"
)

// VolumeSourceType is a type standing for enumeration of ConfigMap and Secret
type VolumeSourceType int

// Enumeration of volume source types: ConfigMap or Secret
const (
	ConfigMapVolumeSourceType VolumeSourceType = iota
	SecretVolumeSourceType
	EmptyDirVolumeSourceType
	PVCVolumeSourceType
	PortFormatErr = "the port specification '%s' is not valid. Please provide in the format 'NAME:PORT', where 'NAME' is optional. Examples: '--port h2c:8080' , '--port 8080'."
)

type MountInfo struct {
	VolumeName string
	SubPath    string
}

func (vt VolumeSourceType) String() string {
	names := [...]string{"config-map", "secret"}
	if vt < ConfigMapVolumeSourceType || vt > SecretVolumeSourceType {
		return "unknown"
	}
	return names[vt]
}

func containerOfPodSpec(spec *corev1.PodSpec) *corev1.Container {
	if len(spec.Containers) == 0 {
		newContainer := corev1.Container{}
		spec.Containers = append(spec.Containers, newContainer)
	}
	return &spec.Containers[0]
}

// UpdateEnvVars gives the configuration all the env var values listed in the given map of
// vars.  Does not touch any environment variables not mentioned, but it can add
// new env vars and change the values of existing ones.
func UpdateEnvVars(spec *corev1.PodSpec, allArgs []string,
	envToUpdate *util.OrderedMap, envToRemove []string,
	envValueFromToUpdate *util.OrderedMap, envValueFromToRemove []string,
	envFileName string, envValueFileToUpdate *util.OrderedMap, envValueFileToRemove []string,
) error {
	container := containerOfPodSpec(spec)

	allEnvsToUpdate := util.NewOrderedMap()

	envIterator := envToUpdate.Iterator()
	envValueFromIterator := envValueFromToUpdate.Iterator()
	envValueFileIterator := envValueFileToUpdate.Iterator()

	envKey, envValue, envExists := envIterator.NextString()
	envValueFromKey, envValueFromValue, envValueFromExists := envValueFromIterator.NextString()
	envValueFileKey, envValueFileValue, envValueFileExists := envValueFileIterator.NextString()
	for _, arg := range allArgs {
		// envs are stored as NAME=value
		if envExists && isValidEnvArg(arg, envKey, envValue) {
			allEnvsToUpdate.Set(envKey, corev1.EnvVar{
				Name:  envKey,
				Value: envValue,
			})
			envKey, envValue, envExists = envIterator.NextString()
		} else if envValueFromExists && isValidEnvValueFromArg(arg, envValueFromKey, envValueFromValue) {
			// envs are stored as NAME=secret:sercretName:key or NAME=config-map:cmName:key
			envVarSource, err := createEnvVarSource(envValueFromValue)
			if err != nil {
				return err
			}
			allEnvsToUpdate.Set(envValueFromKey, corev1.EnvVar{
				Name:      envValueFromKey,
				ValueFrom: envVarSource,
			})
			envValueFromKey, envValueFromValue, envValueFromExists = envValueFromIterator.NextString()
		} else if envValueFileExists && isValidEnvValueFileArg(arg, envFileName) {
			for envValueFileExists {
				allEnvsToUpdate.Set(envValueFileKey, corev1.EnvVar{
					Name:  envValueFileKey,
					Value: envValueFileValue,
				})
				envValueFileKey, envValueFileValue, envValueFileExists = envValueFileIterator.NextString()
			}
		}
	}

	updated := updateEnvVarsFromMap(container.Env, allEnvsToUpdate)
	updated = removeEnvVars(updated, append(envToRemove, envValueFromToRemove...))

	container.Env = updated

	return nil
}

// isValidEnvArg checks that the input arg is a valid argument for specifying env value,
// ie. stored as NAME=value
func isValidEnvArg(arg, envKey, envValue string) bool {
	return strings.HasPrefix(arg, envKey+"="+envValue) || strings.HasPrefix(arg, "-e="+envKey+"="+envValue) || strings.HasPrefix(arg, "--env="+envKey+"="+envValue)
}

// isValidEnvValueFromArg checks that the input arg is a valid argument for specifying env from value,
// ie. stored as NAME=secret:sercretName:key or NAME=config-map:cmName:key
func isValidEnvValueFromArg(arg, envValueFromKey, envValueFromValue string) bool {
	return strings.HasPrefix(arg, envValueFromKey+"="+envValueFromValue) || strings.HasPrefix(arg, "--env-value-from="+envValueFromKey+"="+envValueFromValue)
}

// isValidEnvValueFileArg checks that the input arg is a valid argument for specifying env from value,
// ie. stored as NAME=secret:sercretName:key or NAME=config-map:cmName:key
func isValidEnvValueFileArg(arg, envFileName string) bool {
	return strings.HasPrefix(arg, envFileName) || strings.HasPrefix(arg, "--env-file="+envFileName)
}

// UpdateEnvFrom updates envFrom
func UpdateEnvFrom(spec *corev1.PodSpec, toUpdate []string, toRemove []string) error {
	container := containerOfPodSpec(spec)
	envFrom, err := updateEnvFrom(container.EnvFrom, toUpdate)
	if err != nil {
		return err
	}
	container.EnvFrom, err = removeEnvFrom(envFrom, toRemove)
	return err
}

// UpdateVolumeMountsAndVolumes updates the configuration for volume mounts and volumes.
func UpdateVolumeMountsAndVolumes(spec *corev1.PodSpec,
	mountsToUpdate *util.OrderedMap, mountsToRemove []string, volumesToUpdate *util.OrderedMap, volumesToRemove []string) error {
	container := containerOfPodSpec(spec)

	volumeSourceInfoByName, mountsToUpdate, err := reviseVolumeInfoAndMountsToUpdate(mountsToUpdate, volumesToUpdate)
	if err != nil {
		return err
	}

	volumes, err := updateVolumesFromMap(spec.Volumes, volumeSourceInfoByName)
	if err != nil {
		return err
	}

	volumeMounts, err := updateVolumeMountsFromMap(container.VolumeMounts, mountsToUpdate, volumes)
	if err != nil {
		return err
	}

	volumesToRemove = reviseVolumesToRemove(container.VolumeMounts, volumesToRemove, mountsToRemove)

	container.VolumeMounts = removeVolumeMounts(volumeMounts, mountsToRemove)
	spec.Volumes, err = removeVolumes(volumes, volumesToRemove, container.VolumeMounts)

	return err
}

// UpdateImage a given image
func UpdateImage(spec *corev1.PodSpec, image string) error {
	// When not setting the image to a digest, add the user image annotation.
	container := containerOfPodSpec(spec)
	container.Image = image
	return nil
}

// UpdateContainerCommand updates container with a given argument
func UpdateContainerCommand(spec *corev1.PodSpec, command []string) error {
	container := containerOfPodSpec(spec)
	container.Command = command
	return nil
}

// UpdateContainerArg updates container with a given argument
func UpdateContainerArg(spec *corev1.PodSpec, arg []string) error {
	container := containerOfPodSpec(spec)
	container.Args = arg
	return nil
}

// UpdateContainerPort updates container with a given name:port
func UpdateContainerPort(spec *corev1.PodSpec, port string) error {
	container := containerOfPodSpec(spec)

	var containerPort int64
	var name string
	var err error

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

// UpdateUser updates container with a given user id
func UpdateUser(spec *corev1.PodSpec, user int64) error {
	container := containerOfPodSpec(spec)
	container.SecurityContext = &corev1.SecurityContext{
		RunAsUser: &user,
	}
	return nil
}

// UpdateResources updates container resources for given revision spec
func UpdateResources(spec *corev1.PodSpec, resources corev1.ResourceRequirements, requestsToRemove, limitsToRemove []string) error {
	container := containerOfPodSpec(spec)

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

// UpdateServiceAccountName updates the service account name used for the corresponding knative service
func UpdateServiceAccountName(spec *corev1.PodSpec, serviceAccountName string) {
	serviceAccountName = strings.TrimSpace(serviceAccountName)
	spec.ServiceAccountName = serviceAccountName
}

// UpdateImagePullSecrets updates the image pull secrets used for the corresponding knative service
func UpdateImagePullSecrets(spec *corev1.PodSpec, pullsecrets string) {
	pullsecrets = strings.TrimSpace(pullsecrets)
	if pullsecrets == "" {
		spec.ImagePullSecrets = nil
	} else {
		spec.ImagePullSecrets = []corev1.LocalObjectReference{{
			Name: pullsecrets,
		}}
	}
}

// UpdateContainers updates the containers array with additional ones provided from file or os.Stdin
func UpdateContainers(spec *corev1.PodSpec, containers []corev1.Container) {
	var matched []string
	if len(spec.Containers) == 1 {
		spec.Containers = append(spec.Containers, containers...)
	} else {
		for i, container := range spec.Containers {
			for j, toUpdate := range containers {
				if container.Name == toUpdate.Name {

					spec.Containers[i] = containers[j]

					matched = append(matched, toUpdate.Name)
				}
			}
		}
		for _, container := range containers {
			if !util.SliceContainsIgnoreCase(matched, container.Name) {
				spec.Containers = append(spec.Containers, container)
			}
		}
	}
}

// UpdateLivenessProbe updates container liveness probe based on provided string
func UpdateLivenessProbe(spec *corev1.PodSpec, probeString string) error {
	c := containerOfPodSpec(spec)
	handler, err := resolveProbeHandler(probeString)
	if err != nil {
		return err
	}
	if c.LivenessProbe == nil {
		c.LivenessProbe = &corev1.Probe{}
	}
	c.LivenessProbe.ProbeHandler = *handler
	return nil
}

// UpdateLivenessProbeOpts updates container liveness probe commons options based on provided string
func UpdateLivenessProbeOpts(spec *corev1.PodSpec, probeString string) error {
	c := containerOfPodSpec(spec)
	if c.LivenessProbe == nil {
		c.LivenessProbe = &corev1.Probe{}
	}
	err := resolveProbeOptions(c.LivenessProbe, probeString)
	if err != nil {
		return err
	}
	return nil
}

// UpdateReadinessProbe updates container readiness probe based on provided string
func UpdateReadinessProbe(spec *corev1.PodSpec, probeString string) error {
	c := containerOfPodSpec(spec)
	handler, err := resolveProbeHandler(probeString)
	if err != nil {
		return err
	}
	if c.ReadinessProbe == nil {
		c.ReadinessProbe = &corev1.Probe{}
	}
	c.ReadinessProbe.ProbeHandler = *handler
	return nil
}

// UpdateReadinessProbeOpts updates container readiness probe commons options based on provided string
func UpdateReadinessProbeOpts(spec *corev1.PodSpec, probeString string) error {
	c := containerOfPodSpec(spec)
	if c.ReadinessProbe == nil {
		c.ReadinessProbe = &corev1.Probe{}
	}
	err := resolveProbeOptions(c.ReadinessProbe, probeString)
	if err != nil {
		return err
	}
	return nil
}

// UpdateImagePullPolicy updates the pull policy for the given revision template
func UpdateImagePullPolicy(spec *corev1.PodSpec, imagePullPolicy string) error {
	container := containerOfPodSpec(spec)

	if !isValidPullPolicy(imagePullPolicy) {
		return fmt.Errorf("invalid --pull-policy %s. Valid arguments (case insensitive): Always | Never | IfNotPresent", imagePullPolicy)
	}
	container.ImagePullPolicy = getPolicy(imagePullPolicy)
	return nil
}

func getPolicy(policy string) v1.PullPolicy {
	var ret v1.PullPolicy
	switch strings.ToLower(policy) {
	case "always":
		ret = v1.PullAlways
	case "ifnotpresent":
		ret = v1.PullIfNotPresent
	case "never":
		ret = v1.PullNever
	}
	return ret
}

func isValidPullPolicy(policy string) bool {
	validPolicies := []string{string(v1.PullAlways), string(v1.PullNever), string(v1.PullIfNotPresent)}
	return util.SliceContainsIgnoreCase(validPolicies, policy)
}

// =======================================================================================
func updateEnvVarsFromMap(env []corev1.EnvVar, toUpdate *util.OrderedMap) []corev1.EnvVar {
	updated := sets.NewString()

	for i := range env {
		object, present := toUpdate.Get(env[i].Name)
		if present {
			env[i] = object.(corev1.EnvVar)
			updated.Insert(env[i].Name)
		}
	}
	it := toUpdate.Iterator()
	for name, envVar, ok := it.Next(); ok; name, envVar, ok = it.Next() {
		if !updated.Has(name) {
			env = append(env, envVar.(corev1.EnvVar))
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

func createEnvVarSource(spec string) (*corev1.EnvVarSource, error) {
	slices := strings.SplitN(spec, ":", 3)
	if len(slices) != 3 {
		return nil, fmt.Errorf("argument requires a value in form \"resourceType:name:key\" where \"resourceType\" can be one of \"config-map\" (\"cm\") or \"secret\" (\"sc\"); got %q", spec)
	}

	typeString := strings.TrimSpace(slices[0])
	sourceName := strings.TrimSpace(slices[1])
	sourceKey := strings.TrimSpace(slices[2])

	var sourceType string
	envVarSource := corev1.EnvVarSource{}

	switch typeString {
	case "config-map", "cm":
		sourceType = "ConfigMap"
		envVarSource.ConfigMapKeyRef = &corev1.ConfigMapKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: sourceName,
			},
			Key: sourceKey}
	case "secret", "sc":
		sourceType = "Secret"
		envVarSource.SecretKeyRef = &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: sourceName,
			},
			Key: sourceKey}
	default:
		return nil, fmt.Errorf("unsupported env source type \"%q\"; supported source types are \"config-map\" (\"cm\") and \"secret\" (\"sc\")", slices[0])
	}

	if len(sourceName) == 0 {
		return nil, fmt.Errorf("the name of %s cannot be an empty string", sourceType)
	}

	if len(sourceKey) == 0 {
		return nil, fmt.Errorf("the key referenced by resource %s \"%s\" cannot be an empty string", sourceType, sourceName)
	}

	return &envVarSource, nil
}

// =======================================================================================
func updateEnvFrom(envFromSources []corev1.EnvFromSource, toUpdate []string) ([]corev1.EnvFromSource, error) {
	existingNameSet := make(map[string]bool)

	for i := range envFromSources {
		envSrc := &envFromSources[i]
		if canonicalName, err := getCanonicalNameFromEnvFromSource(envSrc); err == nil {
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
	case EmptyDirVolumeSourceType:
		volume.EmptyDir = &corev1.EmptyDirVolumeSource{Medium: corev1.StorageMedium(info.emptyDirMemoryType), SizeLimit: info.emptyDirSize}
	case PVCVolumeSourceType:
		volume.PersistentVolumeClaim = &corev1.PersistentVolumeClaimVolumeSource{ClaimName: info.volumeSourceName}
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
		mountInfo, present := toUpdate.Get(volumeMount.MountPath)

		if present {
			volumeMountInfo := mountInfo.(*MountInfo)
			name := volumeMountInfo.VolumeName
			if !existsVolumeNameInVolumes(name, volumes) {
				return nil, fmt.Errorf("There is no volume matched with %q", name)
			}
			volumeMount.ReadOnly = isReadOnlyVolume(name, volumes)
			volumeMount.Name = name
			volumeMount.SubPath = volumeMountInfo.SubPath
			set[volumeMount.MountPath] = true
		}
	}

	it := toUpdate.Iterator()
	for mountPath, mountInfo, ok := it.Next(); ok; mountPath, mountInfo, ok = it.Next() {
		volumeMountInfo := mountInfo.(*MountInfo)
		name := volumeMountInfo.VolumeName
		if !set[mountPath] {
			volumeMounts = append(volumeMounts, corev1.VolumeMount{
				Name:      name,
				ReadOnly:  isReadOnlyVolume(name, volumes),
				MountPath: mountPath,
				SubPath:   volumeMountInfo.SubPath,
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
	volumeSourceType   VolumeSourceType
	volumeSourceName   string
	emptyDirMemoryType string
	emptyDirSize       *resource.Quantity
}

func newVolumeSourceInfoWithSpecString(spec string) (*volumeSourceInfo, error) {
	slices := strings.SplitN(spec, ":", 3)
	if len(slices) < 2 {
		return nil, fmt.Errorf("argument requires a value that contains the : character; got %q, %q", spec, slices)
	}

	if len(slices) == 2 {
		var volumeSourceType VolumeSourceType

		typeString := strings.TrimSpace(slices[0])
		volumeSourceName := strings.TrimSpace(slices[1])

		switch typeString {
		case "config-map", "cm":
			volumeSourceType = ConfigMapVolumeSourceType
		case "secret", "sc":
			volumeSourceType = SecretVolumeSourceType
		case "emptyDir", "ed":
			volumeSourceType = EmptyDirVolumeSourceType
		case "persistentVolumeClaim", "pvc":
			volumeSourceType = PVCVolumeSourceType
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
	} else {
		typeString := strings.TrimSpace(slices[0])
		switch typeString {
		case "config-map", "cm", "secret", "sc", "persistentVolumeClaim", "pvc":
			return nil, fmt.Errorf("incorrect mount details for type %q", typeString)
		case "emptyDir", "ed":
			volName := slices[1]
			edType, edSize, err := getEmptyDirTypeAndSize(slices[2])
			if err != nil {
				return nil, err
			}
			return &volumeSourceInfo{
				volumeSourceType:   EmptyDirVolumeSourceType,
				volumeSourceName:   volName,
				emptyDirMemoryType: edType,
				emptyDirSize:       edSize,
			}, nil
		default:
			return nil, fmt.Errorf("unsupported volume type \"%q\"; supported volume types are \"config-map or cm\", \"secret or sc\", \"volume or vo\", and \"emptyDir or ed\"", slices[0])
		}

	}
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

func isReadOnlyVolume(volumeName string, volumes []corev1.Volume) bool {
	for _, volume := range volumes {
		if volume.Name == volumeName {
			return volume.EmptyDir == nil
		}
	}
	return true
}

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

// =======================================================================================

func getMountInfo(volume string) *MountInfo {
	slices := strings.SplitN(volume, "/", 2)
	if len(slices) == 1 || slices[1] == "" {
		return &MountInfo{
			VolumeName: slices[0],
		}
	}
	return &MountInfo{
		VolumeName: slices[0],
		SubPath:    slices[1],
	}
}

func reviseVolumeInfoAndMountsToUpdate(mountsToUpdate *util.OrderedMap, volumesToUpdate *util.OrderedMap) (*util.OrderedMap, *util.OrderedMap, error) {
	volumeSourceInfoByName := util.NewOrderedMap() //make(map[string]*volumeSourceInfo)
	mountsToUpdateRevised := util.NewOrderedMap()  //make(map[string]string)

	it := mountsToUpdate.Iterator()
	for path, value, ok := it.NextString(); ok; path, value, ok = it.NextString() {
		// slices[0] -> config-map, cm, secret, sc, volume, or vo
		// slices[1] -> secret, config-map, or volume name
		slices := strings.SplitN(value, ":", 2)
		if len(slices) == 1 {
			mountInfo := getMountInfo(slices[0])
			mountsToUpdateRevised.Set(path, mountInfo)
		} else {
			switch volumeType := slices[0]; volumeType {
			case "config-map", "cm":
				generatedName := util.GenerateVolumeName(path)
				mountInfo := getMountInfo(slices[1])
				volumeSourceInfoByName.Set(generatedName, &volumeSourceInfo{
					volumeSourceType: ConfigMapVolumeSourceType,
					volumeSourceName: mountInfo.VolumeName,
				})
				mountInfo.VolumeName = generatedName
				mountsToUpdateRevised.Set(path, mountInfo)
			case "secret", "sc":
				generatedName := util.GenerateVolumeName(path)
				mountInfo := getMountInfo(slices[1])
				volumeSourceInfoByName.Set(generatedName, &volumeSourceInfo{
					volumeSourceType: SecretVolumeSourceType,
					volumeSourceName: mountInfo.VolumeName,
				})
				mountInfo.VolumeName = generatedName
				mountsToUpdateRevised.Set(path, mountInfo)
			case "emptyDir", "ed":
				generatedName := util.GenerateVolumeName(path)
				mountInfo := getMountInfo(slices[1])
				volumeSourceInfoByName.Set(generatedName, &volumeSourceInfo{
					volumeSourceType:   EmptyDirVolumeSourceType,
					volumeSourceName:   slices[1],
					emptyDirMemoryType: "",
				})
				mountInfo.VolumeName = generatedName
				mountsToUpdateRevised.Set(path, mountInfo)
			case "persistentVolumeClaim", "pvc":
				generatedName := util.GenerateVolumeName(path)
				mountInfo := getMountInfo(slices[1])
				volumeSourceInfoByName.Set(generatedName, &volumeSourceInfo{
					volumeSourceType: PVCVolumeSourceType,
					volumeSourceName: mountInfo.VolumeName,
				})
				mountInfo.VolumeName = generatedName
				mountsToUpdateRevised.Set(path, mountInfo)
			default:
				return nil, nil, fmt.Errorf("unsupported volume type \"%q\"; supported volume types are \"config-map or cm\", \"secret or sc\", \"volume or vo\", and \"emptyDir or ed\"", slices[0])
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

func getEmptyDirTypeAndSize(value string) (string, *resource.Quantity, error) {
	slices := strings.SplitN(value, ",", 2)
	formatErr := fmt.Errorf("incorrect format to specify emptyDir type")
	repeatErrStr := "cannot repeat the key %q"
	var dirType string
	var size *resource.Quantity
	switch len(slices) {
	case 0:
		return "", nil, nil
	case 1:
		typeSizeSlices := strings.SplitN(slices[0], "=", 2)
		if len(typeSizeSlices) < 2 {
			return "", nil, formatErr
		}
		switch strings.ToLower(typeSizeSlices[0]) {
		case "type":
			dirType = typeSizeSlices[1]
		case "size":
			quantity, err := resource.ParseQuantity(typeSizeSlices[1])
			if err != nil {
				return "", nil, formatErr
			}
			size = &quantity
		default:
			return "", nil, formatErr
		}
	case 2:
		for _, slice := range slices {
			typeSizeSlices := strings.SplitN(slice, "=", 2)
			if len(typeSizeSlices) < 2 {
				return "", nil, formatErr
			}
			switch strings.ToLower(typeSizeSlices[0]) {
			case "type":
				if dirType != "" {
					return "", nil, fmt.Errorf(repeatErrStr, "type")
				}
				dirType = typeSizeSlices[1]
			case "size":
				if size != nil {
					return "", nil, fmt.Errorf(repeatErrStr, "size")
				}
				quantity, err := resource.ParseQuantity(typeSizeSlices[1])
				if err != nil {
					return "", nil, formatErr
				}
				size = &quantity
			default:
				return "", nil, formatErr
			}
		}
	}
	return dirType, size, nil
}

func reviseVolumesToRemove(volumeMounts []corev1.VolumeMount, volumesToRemove []string, mountsToRemove []string) []string {
	for _, pathToRemove := range mountsToRemove {
		for _, volumeMount := range volumeMounts {
			if volumeMount.MountPath == pathToRemove && volumeMount.Name == util.GenerateVolumeName(pathToRemove) {
				volumesToRemove = append(volumesToRemove, volumeMount.Name)
			}
		}
	}
	return volumesToRemove
}

func decodeContainersFromFile(filename string) (*corev1.PodSpec, error) {
	var f *os.File
	var err error
	if filename == "-" {
		f = os.Stdin
	} else {
		f, err = os.Open(filename)
		if err != nil {
			return nil, err
		}
	}
	podSpec := &corev1.PodSpec{}
	decoder := yaml.NewYAMLOrJSONDecoder(f, 512)
	if err = decoder.Decode(podSpec); err != nil {
		return nil, err
	}
	return podSpec, nil
}

// =======================================================================================
// Probes

// resolveProbe parses probes as a string
// It's split into two functions:
//   - resolveProbeOptions() -> common probe opts
//   - resolveProbeHandler() -> probe handler [HTTPGet, Exec, TCPSocket]
// Format:
//	- [http,https]:host:port:path
//	- exec:cmd,cmd,...
//  - tcp:host:port
// Common opts (comma separated, case insensitive):
//	- InitialDelaySeconds=<int_value>,FailureThreshold=<int_value>,
//  	SuccessThreshold=<int_value>,PeriodSeconds==<int_value>,TimeoutSeconds=<int_value>

// resolveProbeOptions parses probe commons options
func resolveProbeOptions(probe *corev1.Probe, probeString string) error {
	options := strings.Split(probeString, ",")
	mappedOptions, err := util.MapFromArray(options, "=")
	if err != nil {
		return err
	}
	for k, v := range mappedOptions {
		// Trim & verify value is convertible to int
		intValue, err := strconv.ParseInt(strings.TrimSpace(v), 0, 32)
		if err != nil {
			return fmt.Errorf("not a nummeric value for parameter '%s'", k)
		}
		// Lower case param name mapping
		switch strings.TrimSpace(strings.ToLower(k)) {
		case "initialdelayseconds":
			probe.InitialDelaySeconds = int32(intValue)
		case "timeoutseconds":
			probe.TimeoutSeconds = int32(intValue)
		case "periodseconds":
			probe.PeriodSeconds = int32(intValue)
		case "successthreshold":
			probe.SuccessThreshold = int32(intValue)
		case "failurethreshold":
			probe.FailureThreshold = int32(intValue)
		default:
			return fmt.Errorf("not a valid probe parameter name '%s'", k)
		}
	}
	return nil
}

// resolveProbeHandler parses probe handler options
func resolveProbeHandler(probeString string) (*corev1.ProbeHandler, error) {
	if len(probeString) == 0 {
		return nil, fmt.Errorf("no probe parameters detected")
	}
	probeParts := strings.Split(probeString, ":")
	if len(probeParts) > 4 {
		return nil, fmt.Errorf("too many probe parameters provided, please check the format")
	}
	var probeHandler *corev1.ProbeHandler
	switch probeParts[0] {
	case "http", "https":
		if len(probeParts) != 4 {
			return nil, fmt.Errorf("unexpected probe format, please use 'http:host:port:path'")
		}
		handler := corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{},
		}
		if probeParts[0] == "https" {
			handler.HTTPGet.Scheme = v1.URISchemeHTTPS
		}
		handler.HTTPGet.Host = probeParts[1]
		if probeParts[2] != "" {
			// Cosmetic fix to have default 'port: 0' instead of empty string 'port: ""'
			handler.HTTPGet.Port = intstr.Parse(probeParts[2])
		}
		handler.HTTPGet.Path = probeParts[3]

		probeHandler = &handler
	case "exec":
		if len(probeParts) != 2 {
			return nil, fmt.Errorf("unexpected probe format, please use 'exec:<exec_command>[,<exec_command>,...]'")
		}
		if len(probeParts[1]) == 0 {
			return nil, fmt.Errorf("at least one command parameter is required for Exec probe")
		}
		handler := corev1.ProbeHandler{
			Exec: &corev1.ExecAction{},
		}
		cmd := strings.Split(probeParts[1], ",")
		handler.Exec.Command = cmd

		probeHandler = &handler
	case "tcp":
		if len(probeParts) != 3 {
			return nil, fmt.Errorf("unexpected probe format, please use 'tcp:host:port")
		}
		handler := corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{},
		}
		handler.TCPSocket.Host = probeParts[1]
		handler.TCPSocket.Port = intstr.Parse(probeParts[2])

		probeHandler = &handler
	default:
		return nil, fmt.Errorf("unsupported probe type '%s'; supported types: http, https, exec, tcp", probeParts[0])
	}
	return probeHandler, nil
}

// =======================================================================================
