## kn source container create

Create a container source

```
kn source container create NAME --image IMAGE --sink SINK
```

### Examples

```

  # Create a ContainerSource 'src' to start a container with image 'docker.io/sample/image' and send messages to service 'mysvc'
  kn source container create src --image docker.io/sample/image --sink ksvc:mysvc
```

### Options

```
      --arg stringArray              Add argument to the container command. Example: --arg myArg1 --arg --myArg2 --arg myArg3=3. You can use this flag multiple times.
      --cmd stringArray              Specify command to be used as entrypoint instead of default one. Example: --cmd /app/start or --cmd sh --cmd /app/start.sh or --cmd /app/start --arg myArg to pass additional arguments.
      --containers string            Specify path to file including definition for additional containers, alternatively use '-' to read from stdin. Example: --containers ./containers.yaml or --containers -.
  -e, --env stringArray              Environment variable to set. NAME=value; you may provide this flag any number of times to set multiple environment variables. To unset, specify the environment variable name followed by a "-" (e.g., NAME-).
      --env-file string              Path to a file containing environment variables (e.g. --env-file=/home/knative/service1/env).
      --env-from stringArray         Add environment variables from a ConfigMap (prefix cm: or config-map:) or a Secret (prefix secret:). Example: --env-from cm:myconfigmap or --env-from secret:mysecret. You can use this flag multiple times. To unset a ConfigMap/Secret reference, append "-" to the name, e.g. --env-from cm:myconfigmap-.
      --env-value-from stringArray   Add environment variable from a value of key in ConfigMap (prefix cm: or config-map:) or a Secret (prefix sc: or secret:). Example: --env-value-from NAME=cm:myconfigmap:key or --env-value-from NAME=secret:mysecret:key. You can use this flag multiple times. To unset a value from a ConfigMap/Secret key reference, append "-" to the key, e.g. --env-value-from ENV-.
  -h, --help                         help for create
      --image string                 Image to run.
      --limit strings                The resource requirement limits for this Service. For example, 'cpu=100m,memory=256Mi'. You can use this flag multiple times. To unset a resource limit, append "-" to the resource name, e.g. '--limit memory-'.
      --mount stringArray            Mount a ConfigMap (prefix cm: or config-map:), a Secret (prefix secret: or sc:), an EmptyDir (prefix ed: or emptyDir:), a PersistentVolumeClaim (prefix pvc: or persistentVolumeClaim) or an existing Volume (without any prefix) on the specified directory. Example: --mount /mydir=cm:myconfigmap, --mount /mydir=secret:mysecret, --mount /mydir=emptyDir:myvol or --mount /mydir=myvolume. When a configmap or a secret is specified, a corresponding volume is automatically generated. You can specify a volume subpath by following the volume name with slash separated path. Example: --mount /mydir=cm:myconfigmap/subpath/to/be/mounted. You can use this flag multiple times. For unmounting a directory, append "-", e.g. --mount /mydir-, which also removes any auto-generated volume.
  -n, --namespace string             Specify the namespace to operate in.
  -p, --port string                  The port where application listens on, in the format 'NAME:PORT', where 'NAME' is optional. Examples: '--port h2c:8080' , '--port 8080'.
      --probe-liveness string        Add liveness probe to Service deployment. Supported probe types are HTTGet, Exec and TCPSocket. Format: [http,https]:host:port:path;<common_opts>, exec:cmd,cmd,...;<common_opts>, tcp:host:port;<common_opts>. Common opts (comma separated, case insensitive): InitialDelaySeconds=<int_value>, FailureThreshold=<int_value>, SuccessThreshold=<int_value>, PeriodSeconds=<int_value>, TimeoutSeconds=<int_value>
      --probe-readiness string       Add readiness probe to Service deployment. Supported probe types are HTTGet, Exec and TCPSocket. Format: [http,https]:host:port:path;<common_opts>, exec:cmd,cmd,...;<common_opts>, tcp:host:port;<common_opts>. Common opts (comma separated, case insensitive): InitialDelaySeconds=<int_value>, FailureThreshold=<int_value>, SuccessThreshold=<int_value>, PeriodSeconds=<int_value>, TimeoutSeconds=<int_value>
      --pull-policy string           Image pull policy. Valid values (case insensitive): Always | Never | IfNotPresent
      --pull-secret string           Image pull secret to set. An empty argument ("") clears the pull secret. The referenced secret must exist in the service's namespace.
      --request strings              The resource requirement requests for this Service. For example, 'cpu=100m,memory=256Mi'. You can use this flag multiple times. To unset a resource request, append "-" to the resource name, e.g. '--request cpu-'.
      --service-account string       Service account name to set. An empty argument ("") clears the service account. The referenced service account must exist in the service's namespace.
  -s, --sink string                  Addressable sink for events. You can specify a broker, channel, Knative service or URI. Examples: '--sink broker:nest' for a broker 'nest', '--sink channel:pipe' for a channel 'pipe', '--sink ksvc:mysvc:mynamespace' for a Knative service 'mysvc' in another namespace 'mynamespace', '--sink https://event.receiver.uri' for an URI with an 'http://' or 'https://' schema, '--sink ksvc:receiver' or simply '--sink receiver' for a Knative service 'receiver' in the current namespace. If a prefix is not provided, it is considered as a Knative service in the current namespace. If referring to a Knative service in another namespace, 'ksvc:name:namespace' combination must be provided explicitly.
      --user int                     The user ID to run the container (e.g., 1001).
      --volume stringArray           Add a volume from a ConfigMap (prefix cm: or config-map:) a Secret (prefix secret: or sc:), an EmptyDir (prefix ed: or emptyDir:) or a PersistentVolumeClaim (prefix pvc: or persistentVolumeClaim). Example: --volume myvolume=cm:myconfigmap, --volume myvolume=secret:mysecret or --volume emptyDir:myvol:size=1Gi,type=Memory. You can use this flag multiple times. To unset a ConfigMap/Secret reference, append "-" to the name, e.g. --volume myvolume-.
```

### Options inherited from parent commands

```
      --cluster string      name of the kubeconfig cluster to use
      --config string       kn configuration file (default: ~/.config/kn/config.yaml)
      --context string      name of the kubeconfig context to use
      --kubeconfig string   kubectl configuration file (default: ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn source container](kn_source_container.md)	 - Manage container sources

