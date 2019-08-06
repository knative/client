## kn service update

Update a service.

### Synopsis

Update a service.

```
kn service update NAME [flags]
```

### Examples

```

  # Updates a service 'mysvc' with new environment variables
  kn service update mysvc --env KEY1=VALUE1 --env KEY2=VALUE2

  # Update a service 'mysvc' with new port
  kn service update mysvc --port 80

  # Updates a service 'mysvc' with new requests and limits parameters
  kn service update mysvc --requests-cpu 500m --limits-memory 1024Mi
```

### Options

```
      --async                    Update service and don't wait for it to become ready.
      --concurrency-limit int    Hard Limit of concurrent requests to be processed by a single replica.
      --concurrency-target int   Recommendation for when to scale up based on the concurrent number of incoming request. Defaults to --concurrency-limit when given.
  -e, --env stringArray          Environment variable to set. NAME=value; you may provide this flag any number of times to set multiple environment variables.
  -h, --help                     help for update
      --image string             Image to run.
      --limits-cpu string        The limits on the requested CPU (e.g., 1000m).
      --limits-memory string     The limits on the requested memory (e.g., 1024Mi).
      --max-scale int            Maximal number of replicas.
      --min-scale int            Minimal number of replicas.
  -n, --namespace string         List the requested object(s) in given namespace.
  -p, --port int32               The port where application listens on.
      --requests-cpu string      The requested CPU (e.g., 250m).
      --requests-memory string   The requested memory (e.g., 64Mi).
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

