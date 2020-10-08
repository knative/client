## kn broker describe

Describe broker

### Synopsis

Describe broker

```
kn broker describe NAME
```

### Examples

```

  # Describe broker 'mybroker' in the current namespace
  kn broker describe mybroker

  # Describe broker 'mybroker' in the 'myproject' namespace
  kn broker describe mybroker --namespace myproject
```

### Options

```
  -h, --help               help for describe
  -n, --namespace string   Specify the namespace to operate in.
```

### Options inherited from parent commands

```
      --config string       kn configuration file (default: ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl configuration file (default: ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn broker](kn_broker.md)	 - Manage message

