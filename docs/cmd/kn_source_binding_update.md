## kn source binding update

Update a sink binding.

### Synopsis

Update a sink binding.

```
kn source binding update NAME --subject SCHEDULE --sink SINK --ce-override OVERRIDE [flags]
```

### Examples

```

  # Update the subject of a sink binding 'my-binding' to a new cronjob with label selector 'app=ping'  
  kn source binding update my-binding --subject cronjob:batch/v1beta1:app=ping"
```

### Options

```
      --ce-override stringArray   Cloud Event overrides to apply before sending event to sink in the format '--ce-override key=value'. --ce-override can be provide multiple times
  -h, --help                      help for update
  -n, --namespace string          Specify the namespace to operate in.
  -s, --sink string               Addressable sink for events
      --subject string            Subject which emits cloud events. This argument takes format kind:apiVersion:name for named resources or kind:apiVersion:labelKey1=value1,labelKey2=value2 for matching via a label selector
```

### Options inherited from parent commands

```
      --config string       kn config file (default is ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl config file (default is ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn source binding](kn_source_binding.md)	 - Sink binding command group

