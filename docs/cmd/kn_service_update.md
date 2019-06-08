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

  # Updates a service 'mysvc' with new requests and limits parameters
  kn service update mysvc --requests-cpu 500m --limits-memory 1024Mi
```

### Options

```
      --concurrency-limit int    Hard Limit of concurrent requests to be processed by a single replica.
      --concurrency-target int   Recommendation for when to scale up based on the concurrent number of incoming request. Defaults to --concurrency-limit when given.
  -e, --env stringArray          Environment variable to set. NAME=value; you may provide this flag any number of times to set multiple environment variables.
  -h, --help                     help for update
      --image string             Image to run.
      --limits-cpu string        The limits on the requested CPU (e.g., 1000m).
      --limits-memory string     The limits on the requested CPU (e.g., 1024Mi).
      --max-scale int            Maximal number of replicas.
      --min-scale int            Minimal number of replicas.
  -n, --namespace string         List the requested object(s) in given namespace.
      --requests-cpu string      The requested CPU (e.g., 250m).
      --requests-memory string   The requested CPU (e.g., 64Mi).
```

### Options inherited from parent commands

```
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
```

### SEE ALSO

* [kn service](kn_service.md)	 - Service command group

