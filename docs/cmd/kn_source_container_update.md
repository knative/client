## kn source container update

Update a container source

```
kn source container update NAME --image IMAGE
```

### Examples

```

  # Update a ContainerSource 'src' with a different image uri 'docker.io/sample/newimage'
  kn source container update src --image docker.io/sample/newimage
```

### Options

```
      --arg stringArray               Add argument to the container command. Example: --arg myArg1 --arg --myArg2 --arg myArg3=3. You can use this flag multiple times.
      --cmd stringArray               Specify command to be used as entrypoint instead of default one. Example: --cmd /app/start or --cmd sh --cmd /app/start.sh or --cmd /app/start --arg myArg to pass additional arguments.
      --containers string             Specify path to file including definition for additional containers, alternatively use '-' to read from stdin. Example: --containers ./containers.yaml or --containers -.
  -e, --env stringArray               Environment variable to set. NAME=value; you may provide this flag any number of times to set multiple environment variables. To unset, specify the environment variable name followed by a "-" (e.g., NAME-).
      --env-file string               Path to a file containing environment variables (e.g. --env-file=/home/knative/service1/env).
      --env-from stringArray          Add environment variables from a ConfigMap (prefix cm: or config-map:) or a Secret (prefix secret:). Example: --env-from cm:myconfigmap or --env-from secret:mysecret. You can use this flag multiple times. To unset a ConfigMap/Secret reference, append "-" to the name, e.g. --env-from cm:myconfigmap-.
      --env-value-from stringArray    Add environment variable from a value of key in ConfigMap (prefix cm: or config-map:) or a Secret (prefix sc: or secret:). Example: --env-value-from NAME=cm:myconfigmap:key or --env-value-from NAME=secret:mysecret:key. You can use this flag multiple times. To unset a value from a ConfigMap/Secret key reference, append "-" to the key, e.g. --env-value-from ENV-.
  -h, --help                          help for update
      --image string                  Image to run.
      --limit strings                 The resource requirement limits for this Service. For example, 'cpu=100m,memory=256Mi'. You can use this flag multiple times. To unset a resource limit, append "-" to the resource name, e.g. '--limit memory-'.
      --mount stringArray             Mount a ConfigMap (prefix cm: or config-map:), a Secret (prefix secret: or sc:), an EmptyDir (prefix ed: or emptyDir:), a PersistentVolumeClaim (prefix pvc: or persistentVolumeClaim) or an existing Volume (without any prefix) on the specified directory. Example: --mount /mydir=cm:myconfigmap, --mount /mydir=secret:mysecret, --mount /mydir=emptyDir:myvol or --mount /mydir=myvolume. When a configmap or a secret is specified, a corresponding volume is automatically generated. You can mount a volume with readOnly config (true | false) also. Example: --mount /mydir=ed:ed1:readOnly=true. You can specify a volume subpath by following the volume name with slash separated path. Example: --mount /mydir=cm:myconfigmap/subpath/to/be/mounted. You can use this flag multiple times. For unmounting a directory, append "-", e.g. --mount /mydir-, which also removes any auto-generated volume.
  -n, --namespace string              Specify the namespace to operate in.
      --node-affinity strings         Add node affinity to be set - only works if the feature gate is enabled in Knative Serving feature flags configuration. When key, operator, values (whitespace separated) and weight are defined for a type, they will be appended in nodeSelectorTerms in case of Required clause, implying the terms will be ORed, and for Preferred clause, all of them will be added in preferredDuringSchedulingIgnoredDuringExecution. Example: --node-affinity Type="Required",Key="topology.kubernetes.io/zone",Operator="In",Values="antarctica-east1 antarctica-west1" or --node-affinity Type="Preferred",Key="topology.kubernetes.io/zone",Operator="In",Values="antarctica-east1",Weight="1"
      --node-selector stringArray     Add node selector to be set, you may provide this flag any number of times to set multiple node selectors, works if feature flag is enabled in Knative Serving feature flags configuration. Example: --node-selector Disktype="ssd". To unset, specify the key name followed by a "-", example: --node-selector Disktype- .
  -p, --port string                   The port where application listens on, in the format 'NAME:PORT', where 'NAME' is optional. Examples: '--port h2c:8080' , '--port 8080'.
      --probe-liveness string         Add liveness probe to Service deployment. Supported probe types are HTTGet, Exec and TCPSocket. Format: [http,https]:host:port:path, exec:cmd[,cmd,...], tcp:host:port.
      --probe-liveness-opts string    Add common options to liveness probe. Common opts (comma separated, case insensitive): InitialDelaySeconds=<int_value>, FailureThreshold=<int_value>, SuccessThreshold=<int_value>, PeriodSeconds=<int_value>, TimeoutSeconds=<int_value>
      --probe-readiness string        Add readiness probe to Service deployment. Supported probe types are HTTGet, Exec and TCPSocket. Format: [http,https]:host:port:path, exec:cmd[,cmd,...], tcp:host:port.
      --probe-readiness-opts string   Add common options to readiness probe. Common opts (comma separated, case insensitive): InitialDelaySeconds=<int_value>, FailureThreshold=<int_value>, SuccessThreshold=<int_value>, PeriodSeconds=<int_value>, TimeoutSeconds=<int_value>
      --pull-policy string            Image pull policy. Valid values (case insensitive): Always | Never | IfNotPresent
      --pull-secret string            Image pull secret to set. An empty argument ("") clears the pull secret. The referenced secret must exist in the service's namespace.
      --request strings               The resource requirement requests for this Service. For example, 'cpu=100m,memory=256Mi'. You can use this flag multiple times. To unset a resource request, append "-" to the resource name, e.g. '--request cpu-'.
      --security-context string       Predefined security context for the service. Accepted values: 'none' for no security context and 'strict' for dropping all capabilities, running as non-root, and no privilege escalation. (default "none")
      --service-account string        Service account name to set. An empty argument ("") clears the service account. The referenced service account must exist in the service's namespace.
  -s, --sink string                   Addressable sink for events. You can specify a broker, channel, Knative service, Kubernetes service or URI. Examples: '--sink broker:nest' for a broker 'nest', '--sink channel:pipe' for a channel 'pipe', '--sink ksvc:mysvc:mynamespace' for a Knative service 'mysvc' in another namespace 'mynamespace', '--sink https://event.receiver.uri' for an HTTP URI, '--sink ksvc:receiver' or simply '--sink receiver' for a Knative service 'receiver' in the current namespace, '--sink svc:receiver:mynamespace' for a Kubernetes service 'receiver' in the 'mynamespace' namespace, '--sink special.eventing.dev/v1alpha1/channels:pipe' for GroupVersionResource of v1alpha1 'pipe'. If a prefix is not provided, it is considered as a Knative service in the current namespace.
      --toleration strings            Add toleration to be set, works if the feature gate is enabled in Knative Serving feature flags configuration. Example: --tolerations Key="key1",Operator="Equal",Value="value1",Effect="NoSchedule"
      --user int                      The user ID to run the container (e.g., 1001).
      --volume stringArray            Add a volume from a ConfigMap (prefix cm: or config-map:) a Secret (prefix secret: or sc:), an EmptyDir (prefix ed: or emptyDir:) or a PersistentVolumeClaim (prefix pvc: or persistentVolumeClaim). PersistentVolumeClaim only works if the feature gate is enabled in Knative Serving feature flags configuration. Example: --volume myvolume=cm:myconfigmap, --volume myvolume=secret:mysecret or --volume emptyDir:myvol:size=1Gi,type=Memory. You can use this flag multiple times. To unset a ConfigMap/Secret reference, append "-" to the name, e.g. --volume myvolume-.
```

### Options inherited from parent commands

```
      --as string              username to impersonate for the operation
      --as-group stringArray   group to impersonate for the operation, this flag can be repeated to specify multiple groups
      --as-uid string          uid to impersonate for the operation
      --cluster string         name of the kubeconfig cluster to use
      --config string          kn configuration file (default: ~/.config/kn/config.yaml)
      --context string         name of the kubeconfig context to use
      --kubeconfig string      kubectl configuration file (default: ~/.kube/config)
      --log-http               log http traffic
```

### SEE ALSO

* [kn source container](kn_source_container.md)	 - Manage container sources

