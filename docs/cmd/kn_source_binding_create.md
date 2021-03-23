## kn source binding create

Create a sink binding

```
kn source binding create NAME --subject SUBJECT --sink SINK
```

### Examples

```

  # Create a sink binding which connects a deployment 'myapp' with a Knative service 'mysvc'
  kn source binding create my-binding --subject Deployment:apps/v1:myapp --sink ksvc:mysvc
```

### Options

```
      --ce-override stringArray   Cloud Event overrides to apply before sending event to sink. Example: '--ce-override key=value' You may be provide this flag multiple times. To unset, append "-" to the key (e.g. --ce-override key-).
  -h, --help                      help for create
  -n, --namespace string          Specify the namespace to operate in.
  -s, --sink string               Addressable sink for events. You can specify a broker, channel, Knative service or URI. Examples: '--sink broker:nest' for a broker 'nest', '--sink channel:pipe' for a channel 'pipe', '--sink ksvc:mysvc:mynamespace' for a Knative service 'mysvc' in another namespace 'mynamespace', '--sink https://event.receiver.uri' for an URI with an 'http://' or 'https://' schema, '--sink ksvc:receiver' or simply '--sink receiver' for a Knative service 'receiver' in the current namespace. If a prefix is not provided, it is considered as a Knative service in the current namespace. If referring to a Knative service in another namespace, 'ksvc:name:namespace' combination must be provided explicitly.
      --subject string            Subject which emits cloud events. This argument takes format kind:apiVersion:name for named resources or kind:apiVersion:labelKey1=value1,labelKey2=value2 for matching via a label selector
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

* [kn source binding](kn_source_binding.md)	 - Manage sink bindings

