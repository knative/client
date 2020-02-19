## kn source binding delete

Delete a sink binding.

### Synopsis

Delete a sink binding.

```
kn source binding delete NAME [flags]
```

### Examples

```

  # Delete a sink binding with name 'my-binding'
  kn source binding delete my-binding
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

* [kn source binding](kn_source_binding.md)	 - Sink binding command group

