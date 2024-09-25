## kn source apiserver update

Update an api-server source

```
kn source apiserver update NAME
```

### Examples

```

  # Update an ApiServerSource 'k8sevents' with different service account and sink service
  kn source apiserver update k8sevents --service-account newsa --sink ksvc:newsvc
```

### Options

```
      --ce-override stringArray   Cloud Event overrides to apply before sending event to sink. Example: '--ce-override key=value' You may be provide this flag multiple times. To unset, append "-" to the key (e.g. --ce-override key-).
  -h, --help                      help for update
      --mode string               The mode the receive adapter controller runs under:,
                                  "Reference" sends only the reference to the resource,
                                  "Resource" send the full resource. (default "Reference")
  -n, --namespace string          Specify the namespace to operate in.
      --resource stringArray      Specification for which events to listen, in the format Kind:APIVersion:LabelSelector, e.g. "Event:sourcesv1:key=value".
                                  "LabelSelector" is a list of comma separated key value pairs. "LabelSelector" can be omitted, e.g. "Event:sourcesv1".
      --service-account string    Name of the service account to use to run this source
  -s, --sink string               Addressable sink for events. You can specify a broker, channel, Knative service, Kubernetes service or URI. Examples: '--sink broker:nest' for a broker 'nest', '--sink channel:pipe' for a channel 'pipe', '--sink ksvc:mysvc:mynamespace' for a Knative service 'mysvc' in another namespace 'mynamespace', '--sink https://event.receiver.uri' for an HTTP URI, '--sink ksvc:receiver' or simply '--sink receiver' for a Knative service 'receiver' in the current namespace, '--sink svc:receiver:mynamespace' for a Kubernetes service 'receiver' in the 'mynamespace' namespace, '--sink special.eventing.dev/v1alpha1/channels:pipe' for GroupVersionResource of v1alpha1 'pipe'. If a prefix is not provided, it is considered as a Knative service in the current namespace.
```

### Options inherited from parent commands

```
      --as string              username to impersonate for the operation
      --as-group stringArray   group to impersonate for the operation, this flag can be repeated to specify multiple groups
      --as-uid string          uid to impersonate for the operation
      --cluster string         name of the kubeconfig cluster to use
      --config string          kn configuration file (default: ~/.config/kn/config.yaml)
      --context string         name of the kubeconfig context to use
      --kubeconfig string      kubectl configuration file (default: ~/.kube/config)
      --log-http               log http traffic
```

### SEE ALSO

* [kn source apiserver](kn_source_apiserver.md)	 - Manage Kubernetes api-server sources

