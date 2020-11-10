## kn source ping create

Create a ping source

### Synopsis

Create a ping source

```
kn source ping create NAME --sink SINK
```

### Examples

```

  # Create a Ping source 'my-ping' which fires every two minutes and sends '{ value: "hello" }' to service 'mysvc' as a cloudevent
  kn source ping create my-ping --schedule "*/2 * * * *" --data '{ value: "hello" }' --sink ksvc:mysvc
```

### Options

```
      --ce-override stringArray   Cloud Event overrides to apply before sending event to sink. Example: '--ce-override key=value' You may be provide this flag multiple times. To unset, append "-" to the key (e.g. --ce-override key-).
  -d, --data string               Json data to send
  -h, --help                      help for create
  -n, --namespace string          Specify the namespace to operate in.
      --schedule string           Optional schedule specification in crontab format (e.g. '*/2 * * * *' for every two minutes. By default fire every minute.
  -s, --sink string               Addressable sink for events. You can specify a broker, channel, Knative service or URI. Examples: '--sink broker:nest' for a broker 'nest', '--sink channel:pipe' for a channel 'pipe', '--sink https://event.receiver.uri' for an URI with an 'http://' or 'https://' schema, '--sink ksvc:receiver' or simply '--sink receiver' for a Knative service 'receiver'. If a prefix is not provided, it is considered as a Knative service.
```

### Options inherited from parent commands

```
      --config string       kn configuration file (default: ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl configuration file (default: ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn source ping](kn_source_ping.md)	 - Manage ping sources

