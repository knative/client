## kn source apiserver delete

Delete an ApiServer source.

### Synopsis

Delete an ApiServer source.

```
kn source apiserver delete NAME [flags]
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
      --config string                    kn config file (default is $HOME/.kn/config.yaml)
      --kubeconfig string                kubectl config file (default is $HOME/.kube/config)
      --log-http string[="__STDERR__"]   log http traffic to stderr (no argument) or a file (with argument) (default "__NO_LOG__")
```

### SEE ALSO

* [kn source apiserver](kn_source_apiserver.md)	 - Kubernetes API Server Event Source command group

