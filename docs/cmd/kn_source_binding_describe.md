## kn source binding describe

Show details of a sink binding

### Synopsis

Show details of a sink binding

```
kn source binding describe NAME [flags]
```

### Examples

```

  # Describe a sink binding with name 'mysinkbinding'
  kn source binding describe mysinkbinding
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

* [kn source binding](kn_source_binding.md)	 - Sink binding command group

