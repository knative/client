## kn service create

Create a service.

### Synopsis

Create a service.

```
kn service create NAME --image IMAGE [flags]
```

### Examples

```

  # Create a service 'mysvc' using image at dev.local/ns/image:latest
  kn service create mysvc --image dev.local/ns/image:latest

  # Create a service with multiple environment variables
  kn service create mysvc --env KEY1=VALUE1 --env KEY2=VALUE2 --image dev.local/ns/image:latest

  # Create or replace a service 's1' with image dev.local/ns/image:v2 using --force flag
  # if service 's1' doesn't exist, it's just a normal create operation
  kn service create --force s1 --image dev.local/ns/image:v2

  # Create or replace environment variables of service 's1' using --force flag
  kn service create --force s1 --env KEY1=NEW_VALUE1 --env NEW_KEY2=NEW_VALUE2 --image dev.local/ns/image:v1

  # Create service 'mysvc' with port 80
  kn service create mysvc --port 80 --image dev.local/ns/image:latest

  # Create or replace default resources of a service 's1' using --force flag
  # (earlier configured resource requests and limits will be replaced with default)
  # (earlier configured environment variables will be cleared too if any)
  kn service create --force s1 --image dev.local/ns/image:v1

  # Create a service with annotation
  kn service create s1 --image dev.local/ns/image:v3 --annotation sidecar.istio.io/inject=false

  # Create a private service (that is a service with no external endpoint)
  kn service create s1 --image dev.local/ns/image:v3 --cluster-local
```

### Options

```
      --annotation stringArray       Service annotation to set. name=value; you may provide this flag any number of times to set multiple annotations. To unset, specify the annotation name followed by a "-" (e.g., name-).
      --arg stringArray              Add argument to the container command. Example: --arg myArg1 --arg --myArg2 --arg myArg3=3. You can use this flag multiple times.
      --async                        DEPRECATED: please use --no-wait instead. Create service and don't wait for it to be ready.
      --autoscale-window string      Duration to look back for making auto-scaling decisions. The service is scaled to zero if no request was received in during that time. (eg: 10s)
      --cluster-local                Specify that the service be private. (--no-cluster-local will make the service publicly available)
      --cmd string                   Specify command to be used as entrypoint instead of default one. Example: --cmd /app/start or --cmd /app/start --arg myArg to pass aditional arguments.
      --concurrency-limit int        Hard Limit of concurrent requests to be processed by a single replica.
      --concurrency-target int       Recommendation for when to scale up based on the concurrent number of incoming request. Defaults to --concurrency-limit when given.
  -e, --env stringArray              Environment variable to set. NAME=value; you may provide this flag any number of times to set multiple environment variables. To unset, specify the environment variable name followed by a "-" (e.g., NAME-).
      --env-from stringArray         Add environment variables from a ConfigMap (prefix cm: or config-map:) or a Secret (prefix secret:). Example: --env-from cm:myconfigmap or --env-from secret:mysecret. You can use this flag multiple times. To unset a ConfigMap/Secret reference, append "-" to the name, e.g. --env-from cm:myconfigmap-.
      --force                        Create service forcefully, replaces existing service if any.
  -h, --help                         help for create
      --image string                 Image to run.
  -l, --label stringArray            Labels to set for both Service and Revision. name=value; you may provide this flag any number of times to set multiple labels. To unset, specify the label name followed by a "-" (e.g., name-).
      --label-revision stringArray   Revision label to set. name=value; you may provide this flag any number of times to set multiple labels. To unset, specify the label name followed by a "-" (e.g., name-). This flag takes precedence over "label" flag.
      --label-service stringArray    Service label to set. name=value; you may provide this flag any number of times to set multiple labels. To unset, specify the label name followed by a "-" (e.g., name-). This flag takes precedence over "label" flag.
      --limits-cpu string            The limits on the requested CPU (e.g., 1000m).
      --limits-memory string         The limits on the requested memory (e.g., 1024Mi).
      --lock-to-digest               Keep the running image for the service constant when not explicitly specifying the image. (--no-lock-to-digest pulls the image tag afresh with each new revision) (default true)
      --max-scale int                Maximal number of replicas.
      --min-scale int                Minimal number of replicas.
      --mount stringArray            Mount a ConfigMap (prefix cm: or config-map:), a Secret (prefix secret: or sc:), or an existing Volume (without any prefix) on the specified directory. Example: --mount /mydir=cm:myconfigmap, --mount /mydir=secret:mysecret, or --mount /mydir=myvolume. When a configmap or a secret is specified, a corresponding volume is automatically generated. You can use this flag multiple times. For unmounting a directory, append "-", e.g. --mount /mydir-, which also removes any auto-generated volume.
  -n, --namespace string             Specify the namespace to operate in.
      --no-cluster-local             Do not specify that the service be private. (--no-cluster-local will make the service publicly available) (default true)
      --no-lock-to-digest            Do not keep the running image for the service constant when not explicitly specifying the image. (--no-lock-to-digest pulls the image tag afresh with each new revision)
      --no-wait                      Create service and don't wait for it to be ready.
  -p, --port int32                   The port where application listens on.
      --pull-secret string           Image pull secret to set. An empty argument ("") clears the pull secret. The referenced secret must exist in the service's namespace.
      --requests-cpu string          The requested CPU (e.g., 250m).
      --requests-memory string       The requested memory (e.g., 64Mi).
      --revision-name string         The revision name to set. Must start with the service name and a dash as a prefix. Empty revision name will result in the server generating a name for the revision. Accepts golang templates, allowing {{.Service}} for the service name, {{.Generation}} for the generation, and {{.Random [n]}} for n random consonants. (default "{{.Service}}-{{.Random 5}}-{{.Generation}}")
      --service-account string       Service account name to set. An empty argument ("") clears the service account. The referenced service account must exist in the service's namespace.
      --user int                     The user ID to run the container (e.g., 1001).
      --volume stringArray           Add a volume from a ConfigMap (prefix cm: or config-map:) or a Secret (prefix secret: or sc:). Example: --volume myvolume=cm:myconfigmap or --volume myvolume=secret:mysecret. You can use this flag multiple times. To unset a ConfigMap/Secret reference, append "-" to the name, e.g. --volume myvolume-.
      --wait-timeout int             Seconds to wait before giving up on waiting for service to be ready. (default 600)
```

### Options inherited from parent commands

```
      --config string       kn config file (default is ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl config file (default is ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn service](kn_service.md)	 - Service command group

