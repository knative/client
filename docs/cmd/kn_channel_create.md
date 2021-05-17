## kn channel create

Create an event channel

```
kn channel create NAME
```

### Examples

```

  # Create a channel 'pipe' with default setting for channel configuration
  kn channel create pipe

  # Create a channel 'imc1' of type InMemoryChannel using inbuilt alias 'imc'
  kn channel create imc1 --type imc
  # same as above without using inbuilt alias but providing explicit GVK
  kn channel create imc1 --type messaging.knative.dev:v1:InMemoryChannel

  # Create a channel 'k1' of type KafkaChannel
  kn channel create k1 --type messaging.knative.dev:v1alpha1:KafkaChannel
```

### Options

```
  -h, --help               help for create
  -n, --namespace string   Specify the namespace to operate in.
      --type string        Override channel type to create, in the format '--type Group:Version:Kind'. If flag is not specified, it uses default messaging layer settings for channel type, cluster wide or specific namespace. You can configure aliases for channel types in kn config and refer the aliases with this flag. You can also refer inbuilt channel type InMemoryChannel using an alias 'imc' like '--type imc'. Examples: '--type messaging.knative.dev:v1alpha1:KafkaChannel' for specifying explicit Group:Version:Kind.
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

* [kn channel](kn_channel.md)	 - Manage event channels

