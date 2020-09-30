## kn subscription update

Update an event subscription

### Synopsis

Update an event subscription

```
kn subscription update NAME
```

### Examples

```

  # Update a subscription 'sub0' with a subscriber ksvc 'receiver'
  kn subscription update sub0 --sink ksvc:receiver

  # Update a subscription 'sub1' with subscriber ksvc 'mirror', reply to a broker 'nest' and DeadLetterSink to a ksvc 'bucket'
  kn subscription update sub1 --sink mirror --sink-reply broker:nest --sink-dead-letter bucket
```

### Options

```
  -h, --help                      help for update
  -n, --namespace string          Specify the namespace to operate in.
  -s, --sink string               Addressable sink for events. You can specify a broker, Knative service or URI. Examples: '--sink broker:nest' for a broker 'nest', '--sink https://event.receiver.uri' for an URI with an 'http://' or 'https://' schema, '--sink ksvc:receiver' or simply '--sink receiver' for a Knative service 'receiver'. If a prefix is not provided, it is considered as a Knative service.
      --sink-dead-letter string   Addressable sink for events. You can specify a broker, Knative service or URI. Examples: '--sink-dead-letter broker:nest' for a broker 'nest', '--sink-dead-letter https://event.receiver.uri' for an URI with an 'http://' or 'https://' schema, '--sink-dead-letter ksvc:receiver' or simply '--sink-dead-letter receiver' for a Knative service 'receiver'. If a prefix is not provided, it is considered as a Knative service.
      --sink-reply string         Addressable sink for events. You can specify a broker, Knative service or URI. Examples: '--sink-reply broker:nest' for a broker 'nest', '--sink-reply https://event.receiver.uri' for an URI with an 'http://' or 'https://' schema, '--sink-reply ksvc:receiver' or simply '--sink-reply receiver' for a Knative service 'receiver'. If a prefix is not provided, it is considered as a Knative service.
```

### Options inherited from parent commands

```
      --config string       kn configuration file (default: ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl configuration file (default: ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn subscription](kn_subscription.md)	 - Manage event subscriptions

