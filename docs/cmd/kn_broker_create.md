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

```

### Options

```
      --backoff-delay string     The delay before retrying.
      --backoff-policy string    The retry backoff policy (linear, exponential).
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
      --cluster string      name of the kubeconfig cluster to use
      --config string       kn configuration file (default: ~/.config/kn/config.yaml)
      --context string      name of the kubeconfig context to use
      --kubeconfig string   kubectl configuration file (default: ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn broker](kn_broker.md)	 - Manage message brokers

