## kn trigger delete

Delete a trigger

```
kn trigger delete NAME
```

### Examples

```

  # Delete a trigger 'mytrigger' in default namespace
  kn trigger delete mytrigger
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
      --cluster string         name of the kubeconfig cluster to use
      --config string          kn configuration file (default: ~/.config/kn/config.yaml)
      --context string         name of the kubeconfig context to use
      --kubeconfig string      kubectl configuration file (default: ~/.kube/config)
      --log-http               log http traffic
```

### SEE ALSO

* [kn trigger](kn_trigger.md)	 - Manage event triggers

