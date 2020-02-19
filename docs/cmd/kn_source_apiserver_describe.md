## kn source apiserver describe

Show details of an ApiServer source

### Synopsis

Show details of an ApiServer source

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
      --config string                    kn config file (default is $HOME/.kn/config.yaml)
      --kubeconfig string                kubectl config file (default is $HOME/.kube/config)
      --log-http string[="__STDERR__"]   log http traffic to stderr (no argument) or a file (with argument) (default "__NO_LOG__")
```

### SEE ALSO

* [kn source apiserver](kn_source_apiserver.md)	 - Kubernetes API Server Event Source command group

