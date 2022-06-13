## kn service create

Create a service

```
kn service create NAME --image IMAGE
```

### Examples

```

  # Create a service 's0' using image knativesamples/helloworld
  kn service create s0 --image knativesamples/helloworld

  # Create a service with multiple environment variables
  kn service create s1 --env TARGET=v1 --env FROM=examples --image knativesamples/helloworld

  # Create or replace a service using --force flag
  # if service 's1' doesn't exist, it's a normal create operation
  kn service create --force s1 --image knativesamples/helloworld

  # Create or replace environment variables of service 's1' using --force flag
  kn service create --force s1 --env TARGET=force --env FROM=examples --image knativesamples/helloworld

  # Create a service with port 8080
  kn service create s2 --port 8080 --image knativesamples/helloworld

  # Create a service with port 8080 and port name h2c
  kn service create s2 --port h2c:8080 --image knativesamples/helloworld

  # Create or replace default resources of a service 's1' using --force flag
  # (earlier configured resource requests and limits will be replaced with default)
  # (earlier configured environment variables will be cleared too if any)
  kn service create --force s1 --image knativesamples/helloworld

  # Create a service with annotation
  kn service create s3 --image knativesamples/helloworld --annotation sidecar.istio.io/inject=false

  # Create a private service (that is a service with no external endpoint)
  kn service create s1 --image knativesamples/helloworld --cluster-local

  # Create a service with 250MB memory, 200m CPU requests and a GPU resource limit
  # [https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/]
  # [https://kubernetes.io/docs/tasks/manage-gpus/scheduling-gpus/]
  kn service create s4gpu --image knativesamples/hellocuda-go --request memory=250Mi,cpu=200m --limit nvidia.com/gpu=1

  # Create the service in offline mode instead of kubernetes cluster (Beta)
  kn service create gitopstest -n test-ns --image knativesamples/helloworld --target=/user/knfiles
  kn service create gitopstest --image knativesamples/helloworld --target=/user/knfiles/test.yaml
  kn service create gitopstest --image knativesamples/helloworld --target=/user/knfiles/test.json
```

### Options

```
  -a, --annotation stringArray            Annotations to set for both Service and Revision. name=value; you may provide this flag any number of times to set multiple annotations.
      --annotation-revision stringArray   Revision annotation to set. name=value; you may provide this flag any number of times to set multiple annotations. To unset, specify the annotation name followed by a "-" (e.g., name-). This flag takes precedence over the "annotation" flag.
      --annotation-service stringArray    Service annotation to set. name=value; you may provide this flag any number of times to set multiple annotations. To unset, specify the annotation name followed by a "-" (e.g., name-). This flag takes precedence over the "annotation" flag.
      --arg stringArray                   Add argument to the container command. Example: --arg myArg1 --arg --myArg2 --arg myArg3=3. You can use this flag multiple times.
      --cluster-local                     Specify that the service be private. (--no-cluster-local will make the service publicly available)
      --cmd stringArray                   Specify command to be used as entrypoint instead of default one. Example: --cmd /app/start or --cmd sh --cmd /app/start.sh or --cmd /app/start --arg myArg to pass additional arguments.
      --concurrency-limit int             Hard Limit of concurrent requests to be processed by a single replica.
      --containers string                 Specify path to file including definition for additional containers, alternatively use '-' to read from stdin. Example: --containers ./containers.yaml or --containers -.
  -e, --env stringArray                   Environment variable to set. NAME=value; you may provide this flag any number of times to set multiple environment variables.
      --env-file string                   Path to a file containing environment variables (e.g. --env-file=/home/knative/service1/env).
      --env-from stringArray              Add environment variables from a ConfigMap (prefix cm: or config-map:) or a Secret (prefix secret:). Example: --env-from cm:myconfigmap or --env-from secret:mysecret. You can use this flag multiple times.
      --env-value-from stringArray        Add environment variable from a value of key in ConfigMap (prefix cm: or config-map:) or a Secret (prefix sc: or secret:). Example: --env-value-from NAME=cm:myconfigmap:key or --env-value-from NAME=secret:mysecret:key. You can use this flag multiple times.
  -f, --filename string                   Create a service from file. The created service can be further modified by combining with other options. For example, -f /path/to/file --env NAME=value adds also an environment variable.
      --force                             Create service forcefully, replaces existing service if any.
  -h, --help                              help for create
      --image string                      Image to run.
  -l, --label stringArray                 Labels to set for both Service and Revision. name=value; you may provide this flag any number of times to set multiple labels.
      --label-revision stringArray        Revision label to set. name=value; you may provide this flag any number of times to set multiple labels. To unset, specify the label name followed by a "-" (e.g., name-). This flag takes precedence over the "label" flag.
      --label-service stringArray         Service label to set. name=value; you may provide this flag any number of times to set multiple labels. To unset, specify the label name followed by a "-" (e.g., name-). This flag takes precedence over the "label" flag.
      --limit strings                     The resource requirement limits for this Service. For example, 'cpu=100m,memory=256Mi'. You can use this flag multiple times. To unset a resource limit, append "-" to the resource name, e.g. '--limit memory-'.
      --lock-to-digest                    Keep the running image for the service constant when not explicitly specifying the image. (--no-lock-to-digest pulls the image tag afresh with each new revision) (default true)
      --mount stringArray                 Mount a ConfigMap (prefix cm: or config-map:), a Secret (prefix secret: or sc:), an EmptyDir (prefix ed: or emptyDir:)or an existing Volume (without any prefix) on the specified directory. Example: --mount /mydir=cm:myconfigmap, --mount /mydir=secret:mysecret, --mount /mydir=emptyDir:myvol or --mount /mydir=myvolume. When a configmap or a secret is specified, a corresponding volume is automatically generated. You can specify a volume subpath by following the volume name with slash separated path. Example: --mount /mydir=cm:myconfigmap/subpath/to/be/mounted. You can use this flag multiple times. For unmounting a directory, append "-", e.g. --mount /mydir-, which also removes any auto-generated volume.
  -n, --namespace string                  Specify the namespace to operate in.
      --no-cluster-local                  Do not specify that the service be private. (--no-cluster-local will make the service publicly available) (default true)
      --no-lock-to-digest                 Do not keep the running image for the service constant when not explicitly specifying the image. (--no-lock-to-digest pulls the image tag afresh with each new revision)
      --no-wait                           Do not wait for 'service create' operation to be completed.
  -p, --port string                       The port where application listens on, in the format 'NAME:PORT', where 'NAME' is optional. Examples: '--port h2c:8080' , '--port 8080'.
      --pull-policy string                Image pull policy. Valid values (case insensitive): Always | Never | IfNotPresent
      --pull-secret string                Image pull secret to set. An empty argument ("") clears the pull secret. The referenced secret must exist in the service's namespace.
      --request strings                   The resource requirement requests for this Service. For example, 'cpu=100m,memory=256Mi'. You can use this flag multiple times. To unset a resource request, append "-" to the resource name, e.g. '--request cpu-'.
      --revision-name string              The revision name to set. Must start with the service name and a dash as a prefix. Empty revision name will result in the server generating a name for the revision. Accepts golang templates, allowing {{.Service}} for the service name, {{.Generation}} for the generation, and {{.Random [n]}} for n random consonants (e.g. {{.Service}}-{{.Random 5}}-{{.Generation}})
      --scale string                      Set the Minimum and Maximum number of replicas. You can use this flag to set both to a single value, or set a range with min/max values, or set either min or max values without specifying the other. Example: --scale 5 (scale-min = 5, scale-max = 5) or --scale 1..5 (scale-min = 1, scale-max = 5) or --scale 1.. (scale-min = 1, scale-max = unchanged) or --scale ..5 (scale-min = unchanged, scale-max = 5)
      --scale-init int                    Initial number of replicas with which a service starts. Can be 0 or a positive integer.
      --scale-max int                     Maximum number of replicas.
      --scale-metric string               Set the name of the metric the PodAutoscaler should scale on. Example: --scale-metric rps (to scale on rps) or --scale-metric concurrency (to scale on concurrency). The default metric is concurrency.
      --scale-min int                     Minimum number of replicas.
      --scale-target int                  Recommendation for what metric value the PodAutoscaler should attempt to maintain. Use with --scale-metric flag to configure the metric name for which the target value should be maintained. Default metric name is concurrency. The flag defaults to --concurrency-limit when given.
      --scale-utilization int             Percentage of concurrent requests utilization before scaling up. (default 70)
      --scale-window string               Duration to look back for making auto-scaling decisions. The service is scaled to zero if no request was received in during that time. (eg: 10s)
      --service-account string            Service account name to set. An empty argument ("") clears the service account. The referenced service account must exist in the service's namespace.
      --tag strings                       Set tag (format: --tag revisionRef=tagName) where revisionRef can be a revision or '@latest' string representing latest ready revision. This flag can be specified multiple times.
      --target string                     Work on local directory instead of a remote cluster (experimental)
      --timeout int                       Duration in seconds that the request routing layer will wait for a request delivered to a container to begin replying (default 300)
      --user int                          The user ID to run the container (e.g., 1001).
      --volume stringArray                Add a volume from a ConfigMap (prefix cm: or config-map:) a Secret (prefix secret: or sc:), or an EmptyDir (prefix ed: or emptyDir:). Example: --volume myvolume=cm:myconfigmap, --volume myvolume=secret:mysecret or --volume emptyDir:myvol:size=1Gi,type=Memory. You can use this flag multiple times. To unset a ConfigMap/Secret reference, append "-" to the name, e.g. --volume myvolume-.
      --wait                              Wait for 'service create' operation to be completed. (default true)
      --wait-timeout int                  Seconds to wait before giving up on waiting for service to be ready. (default 600)
      --wait-window int                   Seconds to wait for service to be ready after a false ready condition is returned (default 2)
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

* [kn service](kn_service.md)	 - Manage Knative services

