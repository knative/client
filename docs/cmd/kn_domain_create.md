## kn domain create

Create a domain mapping

```
kn domain create NAME
```

### Examples

```

  # Create a domain mappings 'hello.example.com' for Knative service 'hello'
  kn domain create hello.example.com --ref hello
```

### Options

```
  -h, --help               help for create
  -n, --namespace string   Specify the namespace to operate in.
      --ref string         Addressable target reference for Domain Mapping. You can specify a Knative Service name.
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

* [kn domain](kn_domain.md)	 - Manage domain mappings

