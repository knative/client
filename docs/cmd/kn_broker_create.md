## kn broker create

Create a broker

```
kn broker create NAME
```

### Examples

```

  # Create a broker 'mybroker' in the current namespace
  kn broker create mybroker

  # Create a broker 'mybroker' in the 'myproject' namespace and with a broker class of 'Kafka'
  kn broker create mybroker --namespace myproject --class Kafka

  # Create a broker 'mybroker' in the myproject namespace with config referencing a configmap in current namespace
  kn broker create mybroker --namespace myproject --class Kafka --broker-config cm:spec-cm
  OR
  kn broker create mybroker --namespace myproject --class Kafka --broker-config spec-cm

  # Create a broker 'mybroker' in the myproject namespace with config referencing secret named spec-sc in test namespace
  kn broker create mybroker --namespace myproject --class Kafka --broker-config sc:spec-sc:test

  # Create a broker 'mybroker' in the myproject namespace with config referencing RabbitmqCluster mycluster in test namespace
  kn broker create mybroker --namespace myproject --class Kafka --broker-config rabbitmq.com/v1beta1:RabbitmqCluster:mycluster:test

```

### Options

```
      --backoff-delay string     The delay before retrying.
      --backoff-policy string    The retry backoff policy (linear, exponential).
      --broker-config string     Reference to the broker configuration For example, a pointer to a ConfigMap (cm:, configmap:), Secret(sc:, secret:), RabbitmqCluster(rmq:, rabbitmq: rabbitmqcluster:) etc. It should be used in conjunction with --class flag. The format for specifying the object is a colon separated string consisting of at most 4 slices:
                                 Length 1: <object-name> (the object will be assumed to be ConfigMap with the same name)
                                 Length 2: <kind>:<object-name> (the APIVersion will be determined for ConfigMap, Secret, and RabbitmqCluster types)
                                 Length 3: <kind>:<object-name>:<namespace> (the APIVersion will be determined only for ConfigMap, Secret, and RabbitmqCluster types. Otherwise it will be interpreted as:
                                 <apiVersion>:<kind>:<object-name>)
                                 Length 4: <apiVersion>:<kind>:<object-name>:<namespace>
      --class string             Broker class like 'MTChannelBasedBroker' or 'Kafka' (if available).
      --dl-sink string           The sink receiving event that could not be sent to a destination.
  -h, --help                     help for create
  -n, --namespace string         Specify the namespace to operate in.
      --retry int32              The minimum number of retries the sender should attempt when sending an event before moving it to the dead letter sink.
      --retry-after-max string   An optional upper bound on the duration specified in a "Retry-After" header when calculating backoff times for retrying 429 and 503 response codes. Setting the value to zero ("PT0S") can be used to opt-out of respecting "Retry-After" header values altogether. This value only takes effect if "Retry" is configured, and also depends on specific implementations (Channels, Sources, etc.) choosing to provide this capability.
      --timeout string           The timeout of each single request. The value must be greater than 0.
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

* [kn broker](kn_broker.md)	 - Manage message brokers

