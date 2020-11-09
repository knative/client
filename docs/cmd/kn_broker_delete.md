## kn broker delete

Delete a broker

### Synopsis

Delete a broker

```
kn broker delete NAME
```

### Examples

```

  # Delete a broker 'mybroker' in the current namespace
  kn broker create mybroker

  # Delete a broker 'mybroker' in the 'myproject' namespace
  kn broker create mybroker --namespace myproject
```

### Options

```
  -h, --help               help for delete
  -n, --namespace string   Specify the namespace to operate in.
      --no-wait            Do not wait for 'broker delete' operation to be completed. (default true)
      --wait               Wait for 'broker delete' operation to be completed.
      --wait-timeout int   Seconds to wait before giving up on waiting for broker to be deleted. (default 600)
```

### Options inherited from parent commands

```
      --config string       kn configuration file (default: ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl configuration file (default: ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn broker](kn_broker.md)	 - Manage message brokers

