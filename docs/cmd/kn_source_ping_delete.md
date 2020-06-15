## kn source ping delete

Delete a Ping source.

### Synopsis

Delete a Ping source.

```
kn source ping delete NAME [flags]
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
      --config string       kn configuration file (default: ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl configuration file (default: ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn source ping](kn_source_ping.md)	 - Ping source command group

