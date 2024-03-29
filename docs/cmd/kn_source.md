## kn source

Manage event sources

```
kn source SOURCE|COMMAND
```

### Options

```
  -h, --help   help for source
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

* [kn](kn.md)	 - kn manages Knative Serving and Eventing resources
* [kn source apiserver](kn_source_apiserver.md)	 - Manage Kubernetes api-server sources
* [kn source binding](kn_source_binding.md)	 - Manage sink bindings
* [kn source container](kn_source_container.md)	 - Manage container sources
* [kn source list](kn_source_list.md)	 - List event sources
* [kn source list-types](kn_source_list-types.md)	 - List event source types
* [kn source ping](kn_source_ping.md)	 - Manage ping sources

