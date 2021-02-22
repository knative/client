## kn source ping delete

Delete a ping source

```
kn source ping delete NAME
```

### Examples

```

  # Delete a Ping source 'my-ping'
  kn source ping delete my-ping
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

* [kn source ping](kn_source_ping.md)	 - Manage ping sources

