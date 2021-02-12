## kn broker create

Create a broker

```
kn broker create NAME
```

### Examples

```

  # Create a broker 'mybroker' in the current namespace
  kn broker create mybroker

  # Create a broker 'mybroker' in the 'myproject' namespace
  kn broker create mybroker --namespace myproject
```

### Options

```
  -h, --help               help for create
  -n, --namespace string   Specify the namespace to operate in.
```

### Options inherited from parent commands

```
      --config string       kn configuration file (default: ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl configuration file (default: ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn broker](kn_broker.md)	 - Manage message brokers

