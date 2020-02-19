## kn trigger delete

Delete a trigger.

### Synopsis

Delete a trigger.

```
kn trigger delete NAME [flags]
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
      --config string                    kn config file (default is $HOME/.kn/config.yaml)
      --kubeconfig string                kubectl config file (default is $HOME/.kube/config)
      --log-http string[="__STDERR__"]   log http traffic to stderr (no argument) or a file (with argument) (default "__NO_LOG__")
```

### SEE ALSO

* [kn trigger](kn_trigger.md)	 - Trigger command group

