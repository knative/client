## kn eventtype create

Create eventtype

```
kn eventtype create
```

### Examples

```

  # Create eventtype 'myeventtype' of type example.type in the current namespace
  kn eventtype create myeventtype --type example.type

  # Create eventtype 'myeventtype' of type example.type in the 'myproject' namespace
  kn eventtype create myeventtype --namespace myproject -t example.type

```

### Options

```
  -b, --broker string      Cloud Event broker
  -h, --help               help for create
  -n, --namespace string   Specify the namespace to operate in.
      --source string      Cloud Event source
  -t, --type string        Cloud Event type
```

### Options inherited from parent commands

```
      --as string              username to impersonate for the operation
      --as-group stringArray   group to impersonate for the operation, this flag can be repeated to specify multiple groups
      --cluster string         name of the kubeconfig cluster to use
      --config string          kn configuration file (default: ~/.config/kn/config.yaml)
      --context string         name of the kubeconfig context to use
      --kubeconfig string      kubectl configuration file (default: ~/.kube/config)
      --log-http               log http traffic
```

### SEE ALSO

* [kn eventtype](kn_eventtype.md)	 - Manage eventtypes

