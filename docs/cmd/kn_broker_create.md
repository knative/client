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
      --backoff-delay string     Based delay between retries
      --backoff-policy string    Backoff policy for retries, either "linear" or "exponential"
      --class string             Broker class like 'MTChannelBasedBroker' or 'Kafka' (if available)
      --dl-sink string           Reference to a sink for delivering events that can not be sent
  -h, --help                     help for create
  -n, --namespace string         Specify the namespace to operate in.
      --retry int32              Number of retries before sending the event to a dead-letter sink
      --retry-after-max string   Upper bound for a duration specified in an "Retry-After" header (experimental)
      --timeout string           Timeout for a single request
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

