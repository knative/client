## kn source binding delete

Delete a sink binding

```
kn source binding delete NAME
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
      --as string              username to impersonate for the operation
      --as-group stringArray   group to impersonate for the operation, this flag can be repeated to specify multiple groups
      --as-uid string          uid to impersonate for the operation
      --cluster string         name of the kubeconfig cluster to use
      --config string          kn configuration file (default: ~/.config/kn/config.yaml)
      --context string         name of the kubeconfig context to use
      --kubeconfig string      kubectl configuration file (default: ~/.kube/config)
      --log-http               log http traffic
  -q, --quiet-mode             run commands in quiet mode
```

### SEE ALSO

* [kn source binding](kn_source_binding.md)	 - Manage sink bindings

