## kn domain delete

Delete a domain mapping

```
kn domain delete NAME
```

### Examples

```

  # Delete domain mappings 'hello.example.com'
  kn domain delete hello.example.com (Beta)
```

### Options

```
  -h, --help               help for delete
  -n, --namespace string   Specify the namespace to operate in.
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

