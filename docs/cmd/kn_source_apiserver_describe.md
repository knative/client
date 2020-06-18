## kn source apiserver describe

Show details of an api-server source

### Synopsis

Show details of an api-server source

```
kn source apiserver describe NAME
```

### Examples

```

  # Describe an ApiServer source with name 'k8sevents'
  kn source apiserver describe k8sevents
```

### Options

```
  -h, --help               help for describe
  -n, --namespace string   Specify the namespace to operate in.
  -v, --verbose            More output.
```

### Options inherited from parent commands

```
      --config string       kn configuration file (default: ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl configuration file (default: ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn source apiserver](kn_source_apiserver.md)	 - Manage Kubernetes api-server sources

