## kn source apiserver create

Create an ApiServer source.

### Synopsis

Create an ApiServer source.

```
kn source apiserver create NAME --resource RESOURCE --service-account ACCOUNTNAME --sink SINK --mode MODE [flags]
```

### Examples

```

  # Create an ApiServerSource 'k8sevents' which consumes Kubernetes events and sends message to service 'mysvc' as a cloudevent
  kn source apiserver create k8sevents --resource Event:v1 --service-account myaccountname --sink svc:mysvc
```

### Options

```
      --ce-override stringArray   Cloud Event overrides to apply before sending event to sink. Example: '--ce-override key=value' You may be provide this flag multiple times. To unset, append "-" to the key (e.g. --ce-override key-).
  -h, --help                      help for create
      --mode string               The mode the receive adapter controller runs under:,
                                  "Reference" sends only the reference to the resource,
                                  "Resource" send the full resource. (default "Reference")
  -n, --namespace string          Specify the namespace to operate in.
      --resource stringArray      Specification for which events to listen, in the format Kind:APIVersion:LabelSelector, e.g. "Event:v1:key=value".
                                  "LabelSelector" is a list of comma separated key value pairs. "LabelSelector" can be omitted, e.g. "Event:v1".
      --service-account string    Name of the service account to use to run this source
  -s, --sink string               Addressable sink for events
```

### Options inherited from parent commands

```
      --config string       kn config file (default is ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl config file (default is ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn source apiserver](kn_source_apiserver.md)	 - Kubernetes API Server Event Source command group

