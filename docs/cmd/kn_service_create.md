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
```

### Options

```
  -e, --env stringArray          Environment variable to set. NAME=value; you may provide this flag any number of times to set multiple environment variables.
  -h, --help                     help for create
      --image string             Image to run.
      --limits-cpu string        The limits on the requested CPU (e.g., 1000m).
      --limits-memory string     The limits on the requested CPU (e.g., 1024Mi).
  -n, --namespace string         List the requested object(s) in given namespace.
      --requests-cpu string      The requested CPU (e.g., 250m).
      --requests-memory string   The requested CPU (e.g., 64Mi).
```

### Options inherited from parent commands

```
      --config string       config file (default is $HOME/.kn.yaml)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
```

### SEE ALSO

* [kn service](kn_service.md)	 - Service command group

