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
      --config string       kn config file (default is ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl config file (default is ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn source binding](kn_source_binding.md)	 - Sink binding command group

