## kn container add

Add a container

```
kn container add NAME
```

### Examples

```

  The command is Beta and may change in the future releases.

  The 'container add' represents utility command that prints YAML container spec to standard output. It's useful for
  multi-container use cases to create definition with help of standard 'kn' option flags. It accepts all container related
  flag available for 'service create'. The command can be chained through Unix pipes to create multiple containers at once.

  # Add a container 'sidecar' from image 'docker.io/example/sidecar' and print it to standard output (Beta)
  kn container add sidecar --image docker.io/example/sidecar

  # Add command can be chained by standard Unix pipe symbol '|' and passed to 'service add|update|apply' commands (Beta)
  kn container add sidecar --image docker.io/example/sidecar:first | \
  kn container add second --image docker.io/example/sidecar:second | \
  kn service create myksvc --image docker.io/example/my-app:latest --containers -
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
  -h, --help                          help for add
      --image string                  Image to run.
      --limit strings                 The resource requirement limits for this Service. For example, 'cpu=100m,memory=256Mi'. You can use this flag multiple times. To unset a resource limit, append "-" to the resource name, e.g. '--limit memory-'.
      --mount stringArray             Mount a ConfigMap (prefix cm: or config-map:), a Secret (prefix secret: or sc:), an EmptyDir (prefix ed: or emptyDir:), a PersistentVolumeClaim (prefix pvc: or persistentVolumeClaim) or an existing Volume (without any prefix) on the specified directory. Example: --mount /mydir=cm:myconfigmap, --mount /mydir=secret:mysecret, --mount /mydir=emptyDir:myvol or --mount /mydir=myvolume. When a configmap or a secret is specified, a corresponding volume is automatically generated. You can mount a volume with readOnly config (true | false) also. Example: --mount /mydir=ed:ed1:readOnly=true. You can specify a volume subpath by following the volume name with slash separated path. Example: --mount /mydir=cm:myconfigmap/subpath/to/be/mounted. You can use this flag multiple times. For unmounting a directory, append "-", e.g. --mount /mydir-, which also removes any auto-generated volume.
  -p, --port string                   The port where application listens on, in the format 'NAME:PORT', where 'NAME' is optional. Examples: '--port h2c:8080' , '--port 8080'.
      --probe-liveness string         Add liveness probe to Service deployment. Supported probe types are HTTGet, Exec and TCPSocket. Format: [http,https]:host:port:path, exec:cmd[,cmd,...], tcp:host:port.
      --probe-liveness-opts string    Add common options to liveness probe. Common opts (comma separated, case insensitive): InitialDelaySeconds=<int_value>, FailureThreshold=<int_value>, SuccessThreshold=<int_value>, PeriodSeconds=<int_value>, TimeoutSeconds=<int_value>
      --probe-readiness string        Add readiness probe to Service deployment. Supported probe types are HTTGet, Exec and TCPSocket. Format: [http,https]:host:port:path, exec:cmd[,cmd,...], tcp:host:port.
      --probe-readiness-opts string   Add common options to readiness probe. Common opts (comma separated, case insensitive): InitialDelaySeconds=<int_value>, FailureThreshold=<int_value>, SuccessThreshold=<int_value>, PeriodSeconds=<int_value>, TimeoutSeconds=<int_value>
      --pull-policy string            Image pull policy. Valid values (case insensitive): Always | Never | IfNotPresent
      --pull-secret string            Image pull secret to set. An empty argument ("") clears the pull secret. The referenced secret must exist in the service's namespace.
      --request strings               The resource requirement requests for this Service. For example, 'cpu=100m,memory=256Mi'. You can use this flag multiple times. To unset a resource request, append "-" to the resource name, e.g. '--request cpu-'.
      --service-account string        Service account name to set. An empty argument ("") clears the service account. The referenced service account must exist in the service's namespace.
      --user int                      The user ID to run the container (e.g., 1001).
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

* [kn container](kn_container.md)	 - Manage service's containers (experimental)

