## kn source apiserver update

update an ApiServerSource, which watches for Kubernetes events and forwards them to a sink

### Synopsis

update an ApiServerSource, which watches for Kubernetes events and forwards them to a sink

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
      --resource strings         Comma seperate Kind:APIVersion:isController list, e.g. Event:v1:true.
                                 "APIVersion" and "isControler" can be omitted.
                                 "APIVersion" is "v1" by default, "isController" is "false" by default.
      --service-account string   Name of the service account to use to run this source
  -s, --sink string              Addressable sink for events
```

### Options inherited from parent commands

```
      --config string       kn config file (default is $HOME/.kn/config.yaml)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn source apiserver](kn_source_apiserver.md)	 - Kubernetes API Server Event Source command group

