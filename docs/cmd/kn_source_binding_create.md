## kn source binding create

Create a sink binding

### Synopsis

Create a sink binding

```
kn source binding create NAME --subject SUBJECT --sink SINK
```

### Examples

```

  # Create a sink binding which connects a deployment 'myapp' with a Knative service 'mysvc'
  kn source binding create my-binding --subject Deployment:apps/v1:myapp --sink svc:mysvc
```

### Options

```
      --ce-override stringArray   Cloud Event overrides to apply before sending event to sink. Example: '--ce-override key=value' You may be provide this flag multiple times. To unset, append "-" to the key (e.g. --ce-override key-).
  -h, --help                      help for create
  -n, --namespace string          Specify the namespace to operate in.
  -s, --sink string               Addressable sink for events. You can specify broker, service or URI. Examples: '--sink broker:nest' for 'nest' broker, '--sink https://event.receiver.uri' for an URI (identified by 'http' or 'https'), '--sink service:receiver' for 'receiver' knative service, you can also use prefix alias 'svc' for 'service'. If do not specify a prefix, it is considered as knative service, i.e. '--sink receiver' is same as '--sink svc:receiver'.
      --subject string            Subject which emits cloud events. This argument takes format kind:apiVersion:name for named resources or kind:apiVersion:labelKey1=value1,labelKey2=value2 for matching via a label selector
```

### Options inherited from parent commands

```
      --config string       kn configuration file (default: ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl configuration file (default: ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn source binding](kn_source_binding.md)	 - Manage sink bindings

