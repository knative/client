## kn source ping create

Create a Ping source.

### Synopsis

Create a Ping source.

```
kn source ping create NAME --schedule SCHEDULE --sink SINK --data DATA [flags]
```

### Examples

```

  # Create a Ping source 'my-ping' which fires every two minutes and sends '{ value: "hello" }' to service 'mysvc' as a cloudevent
  kn source ping create my-ping --schedule "*/2 * * * *" --data '{ value: "hello" }' --sink svc:mysvc
```

### Options

```
  -d, --data string        Json data to send
  -h, --help               help for create
  -n, --namespace string   Specify the namespace to operate in.
      --schedule string    Optional schedule specification in crontab format (e.g. '*/2 * * * *' for every two minutes. By default fire every minute.
  -s, --sink string        Addressable sink for events
```

### Options inherited from parent commands

```
      --config string       kn config file (default is ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl config file (default is ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn source ping](kn_source_ping.md)	 - Ping source command group

