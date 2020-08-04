## kn trigger update

Update a trigger

### Synopsis

Update a trigger

```
kn trigger update NAME
```

### Examples

```

  # Update the filter which key is 'type' to value 'knative.dev.bar' in a trigger 'mytrigger'
  kn trigger update mytrigger --filter type=knative.dev.bar

  # Remove the filter which key is 'type' from a trigger 'mytrigger'
  kn trigger update mytrigger --filter type-

  # Update the sink of a trigger 'mytrigger' to 'ksvc:new-service'
  kn trigger update mytrigger --sink ksvc:new-service
  
```

### Options

```
      --broker string      Name of the Broker which the trigger associates with. (default "default")
      --filter strings     Key-value pair for exact CloudEvent attribute matching against incoming events, e.g type=dev.knative.foo
  -h, --help               help for update
      --inject-broker      Create new broker with name default through common annotation
  -n, --namespace string   Specify the namespace to operate in.
  -s, --sink string        Addressable sink for events. You can specify a broker, Knative service or URI. Examples: '--sink broker:nest' for a broker 'nest', '--sink https://event.receiver.uri' for an URI with an 'http://' or 'https://' schema, '--sink 'ksvc:receiver' or simply '--sink receiver' for a Knative service 'receiver'. If prefix is not provided, it is considered as a Knative service.
```

### Options inherited from parent commands

```
      --config string       kn configuration file (default: ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl configuration file (default: ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn trigger](kn_trigger.md)	 - Manage event triggers

