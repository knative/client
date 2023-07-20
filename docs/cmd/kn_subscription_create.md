## kn subscription create

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
      --channel string            Specify the channel to subscribe to. For the default channel, just use the name (e.g. 'mychannel'). A mapped channel type like 'imc' can be used as a prefix (e.g. 'imc:mychannel'). Finally you can specify the full coordinates to the referenced channel with Group:Version:Kind:Name (e.g. 'messaging.knative.dev:v1alpha1:KafkaChannel:mychannel').
  -h, --help                      help for create
  -n, --namespace string          Specify the namespace to operate in.
  -s, --sink string               Addressable sink for events. You can specify a broker, channel, Knative service or URI. Examples: '--sink broker:nest' for a broker 'nest', '--sink channel:pipe' for a channel 'pipe', '--sink ksvc:mysvc:mynamespace' for a Knative service 'mysvc' in another namespace 'mynamespace', '--sink https://event.receiver.uri' for an URI with an 'http://' or 'https://' schema, '--sink ksvc:receiver' or simply '--sink receiver' for a Knative service 'receiver' in the current namespace. '--sink special.eventing.dev/v1alpha1/channels:pipe' for GroupVersionResource of v1alpha1 'pipe'. GroupVersionResource requires resource name in a lower case plural form. The inputs are sanitized on best effort basis, but in case of 'resource not found' error it may help to double check and correct the input. If a prefix is not provided, it is considered as a Knative service in the current namespace. If referring to a Knative service in another namespace, 'ksvc:name:namespace' combination must be provided explicitly.
      --sink-dead-letter string   Addressable sink for events. You can specify a broker, channel, Knative service or URI. Examples: '--sink-dead-letter broker:nest' for a broker 'nest', '--sink-dead-letter channel:pipe' for a channel 'pipe', '--sink-dead-letter ksvc:mysvc:mynamespace' for a Knative service 'mysvc' in another namespace 'mynamespace', '--sink-dead-letter https://event.receiver.uri' for an URI with an 'http://' or 'https://' schema, '--sink-dead-letter ksvc:receiver' or simply '--sink-dead-letter receiver' for a Knative service 'receiver' in the current namespace. '--sink-dead-letter special.eventing.dev/v1alpha1/channels:pipe' for GroupVersionResource of v1alpha1 'pipe'. GroupVersionResource requires resource name in a lower case plural form. The inputs are sanitized on best effort basis, but in case of 'resource not found' error it may help to double check and correct the input. If a prefix is not provided, it is considered as a Knative service in the current namespace. If referring to a Knative service in another namespace, 'ksvc:name:namespace' combination must be provided explicitly.
      --sink-reply string         Addressable sink for events. You can specify a broker, channel, Knative service or URI. Examples: '--sink-reply broker:nest' for a broker 'nest', '--sink-reply channel:pipe' for a channel 'pipe', '--sink-reply ksvc:mysvc:mynamespace' for a Knative service 'mysvc' in another namespace 'mynamespace', '--sink-reply https://event.receiver.uri' for an URI with an 'http://' or 'https://' schema, '--sink-reply ksvc:receiver' or simply '--sink-reply receiver' for a Knative service 'receiver' in the current namespace. '--sink-reply special.eventing.dev/v1alpha1/channels:pipe' for GroupVersionResource of v1alpha1 'pipe'. GroupVersionResource requires resource name in a lower case plural form. The inputs are sanitized on best effort basis, but in case of 'resource not found' error it may help to double check and correct the input. If a prefix is not provided, it is considered as a Knative service in the current namespace. If referring to a Knative service in another namespace, 'ksvc:name:namespace' combination must be provided explicitly.
```

### Options inherited from parent commands

```
      --as string              username to impersonate for the operation
      --as-group stringArray   group to impersonate for the operation, this flag can be repeated to specify multiple groups
      --as-uid string          uid to impersonate for the operation
      --cluster string         name of the kubeconfig cluster to use
      --config string          kn configuration file (default: ~/.config/kn/config.yaml)
      --context string         name of the kubeconfig context to use
      --kubeconfig string      kubectl configuration file (default: ~/.kube/config)
      --log-http               log http traffic
```

### SEE ALSO

* [kn subscription](kn_subscription.md)	 - Manage event subscriptions

