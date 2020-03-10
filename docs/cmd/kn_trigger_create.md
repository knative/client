## kn trigger create

Create a trigger

### Synopsis

Create a trigger

```
kn trigger create NAME --broker BROKER --filter KEY=VALUE --sink SINK [flags]
```

### Examples

```

  # Create a trigger 'mytrigger' to declare a subscription to events with attribute 'type=dev.knative.foo' from default broker. The subscriber is service 'mysvc'
  kn trigger create mytrigger --broker default --filter type=dev.knative.foo --sink svc:mysvc
```

### Options

```
      --broker string      Name of the Broker which the trigger associates with. (default "default")
      --filter strings     Key-value pair for exact CloudEvent attribute matching against incoming events, e.g type=dev.knative.foo
  -h, --help               help for create
      --inject-broker      Create new broker with name default through common annotation
  -n, --namespace string   Specify the namespace to operate in.
  -s, --sink string        Addressable sink for events
```

### Options inherited from parent commands

```
      --config string       kn config file (default is ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl config file (default is ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn trigger](kn_trigger.md)	 - Trigger command group

