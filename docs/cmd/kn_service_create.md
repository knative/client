## kn service create

Create a service

### Synopsis

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
```

### Options

```
  -a, --annotation stringArray            Annotations to set for both Service and Revision. name=value; you may provide this flag any number of times to set multiple annotations. To unset, specify the annotation name followed by a "-" (e.g., name-).
      --annotation-revision stringArray   Revision annotation to set. name=value; you may provide this flag any number of times to set multiple annotations. To unset, specify the annotation name followed by a "-" (e.g., name-). This flag takes precedence over the "annotation" flag.
      --annotation-service stringArray    Service annotation to set. name=value; you may provide this flag any number of times to set multiple annotations. To unset, specify the annotation name followed by a "-" (e.g., name-). This flag takes precedence over the "annotation" flag.
      --arg stringArray                   Add argument to the container command. Example: --arg myArg1 --arg --myArg2 --arg myArg3=3. You can use this flag multiple times.
      --async                             DEPRECATED: please use --no-wait instead. Do not wait for 'service create' operation to be completed.
      --autoscale-window string           Duration to look back for making auto-scaling decisions. The service is scaled to zero if no request was received in during that time. (eg: 10s)
      --cluster-local                     Specify that the service be private. (--no-cluster-local will make the service publicly available)
      --cmd string                        Specify command to be used as entrypoint instead of default one. Example: --cmd /app/start or --cmd /app/start --arg myArg to pass aditional arguments.
      --concurrency-limit int             Hard Limit of concurrent requests to be processed by a single replica.
      --concurrency-target int            Recommendation for when to scale up based on the concurrent number of incoming request. Defaults to --concurrency-limit when given.
      --concurrency-utilization int       Percentage of concurrent requests utilization before scaling up. (default 70)
  -e, --env stringArray                   Environment variable to set. NAME=value; you may provide this flag any number of times to set multiple environment variables. To unset, specify the environment variable name followed by a "-" (e.g., NAME-).
      --env-from stringArray              Add environment variables from a ConfigMap (prefix cm: or config-map:) or a Secret (prefix secret:). Example: --env-from cm:myconfigmap or --env-from secret:mysecret. You can use this flag multiple times. To unset a ConfigMap/Secret reference, append "-" to the name, e.g. --env-from cm:myconfigmap-.
  -f, --filename string                   Create a service from file. The created service can be further modified by combining with other options. For example, -f /path/to/file --env NAME=value adds also an environment variable.
      --force                             Create service forcefully, replaces existing service if any.
  -h, --help                              help for create
      --image string                      Image to run.
  -l, --label stringArray                 Labels to set for both Service and Revision. name=value; you may provide this flag any number of times to set multiple labels. To unset, specify the label name followed by a "-" (e.g., name-).
      --label-revision stringArray        Revision label to set. name=value; you may provide this flag any number of times to set multiple labels. To unset, specify the label name followed by a "-" (e.g., name-). This flag takes precedence over the "label" flag.
      --label-service stringArray         Service label to set. name=value; you may provide this flag any number of times to set multiple labels. To unset, specify the label name followed by a "-" (e.g., name-). This flag takes precedence over the "label" flag.
      --limit strings                     The resource requirement limits for this Service. For example, 'cpu=100m,memory=256Mi'. You can use this flag multiple times. To unset a resource limit, append "-" to the resource name, e.g. '--limit memory-'.
      --limits-cpu string                 DEPRECATED: please use --limit instead. The limits on the requested CPU (e.g., 1000m).
      --limits-memory string              DEPRECATED: please use --limit instead. The limits on the requested memory (e.g., 1024Mi).
      --lock-to-digest                    Keep the running image for the service constant when not explicitly specifying the image. (--no-lock-to-digest pulls the image tag afresh with each new revision) (default true)
      --mount stringArray                 Mount a ConfigMap (prefix cm: or config-map:), a Secret (prefix secret: or sc:), or an existing Volume (without any prefix) on the specified directory. Example: --mount /mydir=cm:myconfigmap, --mount /mydir=secret:mysecret, or --mount /mydir=myvolume. When a configmap or a secret is specified, a corresponding volume is automatically generated. You can use this flag multiple times. For unmounting a directory, append "-", e.g. --mount /mydir-, which also removes any auto-generated volume.
  -n, --namespace string                  Specify the namespace to operate in.
      --no-cluster-local                  Do not specify that the service be private. (--no-cluster-local will make the service publicly available) (default true)
      --no-lock-to-digest                 Do not keep the running image for the service constant when not explicitly specifying the image. (--no-lock-to-digest pulls the image tag afresh with each new revision)
      --no-wait                           Do not wait for 'service create' operation to be completed.
  -p, --port string                       The port where application listens on, in the format 'NAME:PORT', where 'NAME' is optional. Examples: '--port h2c:8080' , '--port 8080'.
      --pull-secret string                Image pull secret to set. An empty argument ("") clears the pull secret. The referenced secret must exist in the service's namespace.
      --request strings                   The resource requirement requests for this Service. For example, 'cpu=100m,memory=256Mi'. You can use this flag multiple times. To unset a resource request, append "-" to the resource name, e.g. '--request cpu-'.
      --requests-cpu string               DEPRECATED: please use --request instead. The requested CPU (e.g., 250m).
      --requests-memory string            DEPRECATED: please use --request instead. The requested memory (e.g., 64Mi).
      --revision-name string              The revision name to set. Must start with the service name and a dash as a prefix. Empty revision name will result in the server generating a name for the revision. Accepts golang templates, allowing {{.Service}} for the service name, {{.Generation}} for the generation, and {{.Random [n]}} for n random consonants. (default "{{.Service}}-{{.Random 5}}-{{.Generation}}")
      --scale int                         Minimum and maximum number of replicas.
      --scale-init int                    Initial number of replicas with which a service starts. Can be 0 or a positive integer.
      --scale-max int                     Maximum number of replicas.
      --scale-min int                     Minimum number of replicas.
      --service-account string            Service account name to set. An empty argument ("") clears the service account. The referenced service account must exist in the service's namespace.
      --user int                          The user ID to run the container (e.g., 1001).
      --volume stringArray                Add a volume from a ConfigMap (prefix cm: or config-map:) or a Secret (prefix secret: or sc:). Example: --volume myvolume=cm:myconfigmap or --volume myvolume=secret:mysecret. You can use this flag multiple times. To unset a ConfigMap/Secret reference, append "-" to the name, e.g. --volume myvolume-.
      --wait                              Wait for 'service create' operation to be completed. (default true)
      --wait-timeout int                  Seconds to wait before giving up on waiting for service to be ready. (default 600)
```

### Options inherited from parent commands

```
      --config string       kn configuration file (default: ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl configuration file (default: ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn service](kn_service.md)	 - Manage Knative services

