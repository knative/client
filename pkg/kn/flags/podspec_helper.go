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
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
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
func UpdateEnvVars(spec *corev1.PodSpec,
	allArgs []string, envToUpdate *util.OrderedMap, envToRemove []string, envValueFromToUpdate *util.OrderedMap, envValueFromToRemove []string) error {
	container := containerOfPodSpec(spec)

	allEnvsToUpdate := util.NewOrderedMap()

	envIterator := envToUpdate.Iterator()
	envValueFromIterator := envValueFromToUpdate.Iterator()

	envKey, envValue, envExists := envIterator.NextString()
	envValueFromKey, envValueFromValue, envValueFromExists := envValueFromIterator.NextString()
	for _, arg := range allArgs {
		// envs are stored as NAME=value
		if envExists &&
			(strings.HasPrefix(arg, envKey+"="+envValue) || strings.HasPrefix(arg, "-e="+envKey+"="+envValue) || strings.HasPrefix(arg, "--env="+envKey+"="+envValue)) {
			allEnvsToUpdate.Set(envKey, corev1.EnvVar{
				Name:  envKey,
				Value: envValue,
			})
			envKey, envValue, envExists = envIterator.NextString()
		} else if envValueFromExists &&
			(strings.HasPrefix(arg, envValueFromKey+"="+envValueFromValue) || strings.HasPrefix(arg, "--env-value-from="+envValueFromKey+"="+envValueFromValue)) {
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
		}
	}

	updated := updateEnvVarsFromMap(container.Env, allEnvsToUpdate)
	updated = removeEnvVars(updated, append(envToRemove, envValueFromToRemove...))

	container.Env = updated

	return nil
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
func UpdateContainerCommand(spec *corev1.PodSpec, command string) error {
	container := containerOfPodSpec(spec)
	container.Command = []string{command}
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

// =======================================================================================

func reviseVolumeInfoAndMountsToUpdate(mountsToUpdate *util.OrderedMap, volumesToUpdate *util.OrderedMap) (*util.OrderedMap, *util.OrderedMap, error) {
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
				generatedName := util.GenerateVolumeName(path)
				volumeSourceInfoByName.Set(generatedName, &volumeSourceInfo{
					volumeSourceType: ConfigMapVolumeSourceType,
					volumeSourceName: slices[1],
				})
				mountsToUpdateRevised.Set(path, generatedName)
			case "secret", "sc":
				generatedName := util.GenerateVolumeName(path)
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
			if volumeMount.MountPath == pathToRemove && volumeMount.Name == util.GenerateVolumeName(pathToRemove) {
				volumesToRemove = append(volumesToRemove, volumeMount.Name)
			}
		}
	}
	return volumesToRemove
}
