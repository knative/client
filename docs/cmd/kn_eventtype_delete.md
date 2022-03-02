## kn eventtype delete

Delete eventtype

```
kn eventtype delete
```

### Examples

```

  # Delete eventtype 'myeventtype' in the current namespace
  kn eventtype delete myeventtype

  # Delete eventtype 'myeventtype' in the 'myproject' namespace
  kn eventtype delete myeventtype --namespace myproject

```

### Options

```
  -h, --help               help for delete
  -n, --namespace string   Specify the namespace to operate in.
```

### Options inherited from parent commands

```
      --cluster string      name of the kubeconfig cluster to use
      --config string       kn configuration file (default: ~/.config/kn/config.yaml)
      --context string      name of the kubeconfig context to use
      --kubeconfig string   kubectl configuration file (default: ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn eventtype](kn_eventtype.md)	 - Manage eventtypes

