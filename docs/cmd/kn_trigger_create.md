## kn trigger create

Create a trigger

### Synopsis

Create a trigger

```
kn trigger create NAME --sink SINK
```

### Examples

```

  # Create a trigger 'mytrigger' to declare a subscription to events from default broker. The subscriber is service 'mysvc'
  kn trigger create mytrigger --broker default --sink ksvc:mysvc

  # Create a trigger to filter events with attribute 'type=dev.knative.foo'
  kn trigger create mytrigger --broker default --filter type=dev.knative.foo --sink ksvc:mysvc
```

### Options

```
      --broker string      Name of the Broker which the trigger associates with. (default "default")
      --filter strings     Key-value pair for exact CloudEvent attribute matching against incoming events, e.g type=dev.knative.foo
  -h, --help               help for create
      --inject-broker      Create new broker with name default through common annotation
  -n, --namespace string   Specify the namespace to operate in.
  -s, --sink string        Addressable sink for events. You can specify a broker, Knative service or URI. Examples: '--sink broker:nest' for a broker 'nest', '--sink https://event.receiver.uri' for an URI with an 'http://' or 'https://' schema, '--sink ksvc:receiver' or simply '--sink receiver' for a Knative service 'receiver'. If a prefix is not provided, it is considered as a Knative service.
```

### Options inherited from parent commands

```
      --config string       kn configuration file (default: ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl configuration file (default: ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn trigger](kn_trigger.md)	 - Manage event triggers

