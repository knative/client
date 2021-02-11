## kn source apiserver delete

Delete an api-server source

```
kn source apiserver delete NAME
```

### Examples

```

  # Delete an ApiServerSource 'k8sevents' in default namespace
  kn source apiserver delete k8sevents
```

### Options

```
  -h, --help               help for delete
  -n, --namespace string   Specify the namespace to operate in.
```

### Options inherited from parent commands

```
      --config string       kn configuration file (default: ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl configuration file (default: ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn source apiserver](kn_source_apiserver.md)	 - Manage Kubernetes api-server sources

