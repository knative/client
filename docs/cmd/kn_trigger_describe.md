## kn trigger describe

Describe a trigger.

### Synopsis

Describe a trigger.

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
      --config string       kn config file (default is $HOME/.kn/config.yaml)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn trigger](kn_trigger.md)	 - Trigger command group

