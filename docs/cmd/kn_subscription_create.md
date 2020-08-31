## kn subscription create

Create a subscription

### Synopsis

Create a subscription

```
kn subscription create NAME
```

### Examples

```

  # Create a subscription 'sub0' from InMemoryChannel 'pipe0' to a subscriber ksvc 'receiver'
  kn subscription create sub0 --channel imcv1beta1:pipe0 --sink ksvc:receiver

  # Create a subscription 'sub1' from KafkaChannel 'k1' to ksvc 'mirror', reply to a broker 'nest' and DeadLetterSink to a ksvc 'bucket'
  kn subscription create sub1 --channel messaging.knative.dev:v1alpha1:KafkaChannel:k1 --sink mirror --sink-reply broker:nest --sink-dead-letter bucket
```

### Options

```
      --channel string            Specify the channel to subscribe to, in the format '--channel Group:Version:Kind:Name'. You can use channel type aliases from kn config with this flag. You can also refer inbuilt channel type InMemoryChannel using alias 'imc' like '--type imc:CHANNEL_NAME'. Examples: '--channel messaging.knative.dev:v1alpha1:KafkaChannel:k1' for specifying explicit Group:Version:Kind:Name.
  -h, --help                      help for create
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

