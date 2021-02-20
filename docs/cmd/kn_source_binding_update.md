## kn source binding update

Update a sink binding

```
kn source binding update NAME
```

### Examples

```

  # Update the subject of a sink binding 'my-binding' to a new cronjob with label selector 'app=ping'  
  kn source binding update my-binding --subject cronjob:batch/v1beta1:app=ping"
```

### Options

```
      --ce-override stringArray   Cloud Event overrides to apply before sending event to sink. Example: '--ce-override key=value' You may be provide this flag multiple times. To unset, append "-" to the key (e.g. --ce-override key-).
  -h, --help                      help for update
  -n, --namespace string          Specify the namespace to operate in.
  -s, --sink string               Addressable sink for events. You can specify a broker, channel, Knative service or URI. Examples: '--sink broker:nest' for a broker 'nest', '--sink channel:pipe' for a channel 'pipe', '--sink https://event.receiver.uri' for an URI with an 'http://' or 'https://' schema, '--sink ksvc:receiver' or simply '--sink receiver' for a Knative service 'receiver'. If a prefix is not provided, it is considered as a Knative service.
      --subject string            Subject which emits cloud events. This argument takes format kind:apiVersion:name for named resources or kind:apiVersion:labelKey1=value1,labelKey2=value2 for matching via a label selector
```

### Options inherited from parent commands

```
      --config string       kn configuration file (default: ~/.config/kn/config.yaml)
      --context string      name of the kubeconfig context to use
      --kubeconfig string   kubectl configuration file (default: ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn source binding](kn_source_binding.md)	 - Manage sink bindings

