## kn source apiserver describe

Describe an ApiServer source.

### Synopsis

Describe an ApiServer source.

```
kn source apiserver describe NAME [flags]
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
      --config string       kn config file (default is $HOME/.kn/config.yaml)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn source apiserver](kn_source_apiserver.md)	 - Kubernetes API Server Event Source command group

