## kn trigger describe

Show details of a trigger

### Synopsis

Show details of a trigger

```
kn trigger describe NAME [flags]
```

### Examples

```

  # Describe a trigger with name 'my-trigger'
  kn trigger describe my-trigger
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

* [kn trigger](kn_trigger.md)	 - Trigger command group

