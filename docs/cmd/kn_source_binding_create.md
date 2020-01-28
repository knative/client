## kn source binding create

Create a sink binding source.

### Synopsis

Create a sink binding source.

```
kn source binding create NAME --subject SCHEDULE --sink SINK --ce-override KEY=VALUE [flags]
```

### Examples

```

  # Create a sink binding source, which connects a deployment 'myapp' with a Knative service 'mysvc'
  kn source binding create my-binding --subject "" --sink svc:mysvc
```

### Options

```
      --ce-override stringArray   Cloud Event overrides to apply before sending event to sink in the format '--ce-override key=value'. --ce-override can be provide multiple times
  -h, --help                      help for create
  -n, --namespace string          Specify the namespace to operate in.
  -s, --sink string               Addressable sink for events
      --subject string            Subject which emits cloud events
```

### Options inherited from parent commands

```
      --config string       kn config file (default is $HOME/.kn/config.yaml)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn source binding](kn_source_binding.md)	 - Sink binding command group

