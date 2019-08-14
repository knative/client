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
```

### Options

```
      --async                    Create service and don't wait for it to become ready.
      --concurrency-limit int    Hard Limit of concurrent requests to be processed by a single replica.
      --concurrency-target int   Recommendation for when to scale up based on the concurrent number of incoming request. Defaults to --concurrency-limit when given.
  -e, --env stringArray          Environment variable to set. NAME=value; you may provide this flag any number of times to set multiple environment variables. To unset, specify the environment variable name followed by a "-" (e.g., NAME-).
      --force                    Create service forcefully, replaces existing service if any.
  -h, --help                     help for create
      --image string             Image to run.
  -l, --label stringArray        Service label to set. name=value; you may provide this flag any number of times to set multiple labels. To unset, specify the label name followed by a "-" (e.g., name-).
      --limits-cpu string        The limits on the requested CPU (e.g., 1000m).
      --limits-memory string     The limits on the requested memory (e.g., 1024Mi).
      --max-scale int            Maximal number of replicas.
      --min-scale int            Minimal number of replicas.
  -n, --namespace string         List the requested object(s) in given namespace.
  -p, --port int32               The port where application listens on.
      --requests-cpu string      The requested CPU (e.g., 250m).
      --requests-memory string   The requested memory (e.g., 64Mi).
      --revision-name string     The revision name to set. Must start with the service name and a dash as a prefix. Empty revision name will result in the server generating a name for the revision. Accepts golang templates, allowing {{.Service}} for the service name, {{.Generation}} for the generation, and {{.Random [n]}} for n random consonants. (default "{{.Service}}-{{.Random 5}}-{{.Generation}}")
      --wait-timeout int         Seconds to wait before giving up on waiting for service to be ready. (default 60)
```

### Options inherited from parent commands

```
      --config string       kn config file (default is $HOME/.kn/config.yaml)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn service](kn_service.md)	 - Service command group

