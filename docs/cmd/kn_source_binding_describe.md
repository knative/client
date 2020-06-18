## kn source binding describe

Show details of a sink binding

### Synopsis

Show details of a sink binding

```
kn source binding describe NAME
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
      --config string       kn configuration file (default: ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl configuration file (default: ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn source binding](kn_source_binding.md)	 - Manage sink bindings

