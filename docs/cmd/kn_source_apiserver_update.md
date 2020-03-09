## kn source apiserver update

Update an ApiServer source.

### Synopsis

Update an ApiServer source.

```
kn source apiserver update NAME --resource RESOURCE --service-account ACCOUNTNAME --sink SINK --mode MODE [flags]
```

### Examples

```

  # Update an ApiServerSource 'k8sevents' with different service account and sink service
  kn source apiserver update k8sevents --service-account newsa --sink svc:newsvc
```

### Options

```
  -h, --help                     help for update
      --mode string              The mode the receive adapter controller runs under:,
                                 "Ref" sends only the reference to the resource,
                                 "Resource" send the full resource. (default "Ref")
  -n, --namespace string         Specify the namespace to operate in.
      --resource stringArray     Specification for which events to listen, in the format Kind:APIVersion:isController, e.g. "Event:v1:true".
                                 "isController" can be omitted and is "false" by default, e.g. "Event:v1".
      --service-account string   Name of the service account to use to run this source
  -s, --sink string              Addressable sink for events
```

### Options inherited from parent commands

```
      --config string       kn config file (default is ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl config file (default is ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn source apiserver](kn_source_apiserver.md)	 - Kubernetes API Server Event Source command group

