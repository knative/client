## kn source ping update

Update a Ping source.

### Synopsis

Update a Ping source.

```
kn source ping update NAME --schedule SCHEDULE --sink SERVICE --data DATA [flags]
```

### Examples

```

  # Update the schedule of a Ping source 'my-ping' to fire every minute
  kn source ping update my-ping --schedule "* * * * *"
```

### Options

```
  -d, --data string        Json data to send
  -h, --help               help for update
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

