## kn plugin

Manage kn plugins

### Synopsis

Manage kn plugins

Plugins provide extended functionality that is not part of the core kn command-line distribution.
Please refer to the documentation and examples for more information about how to write your own plugins.

```
kn plugin
```

### Options

```
  -h, --help   help for plugin
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
  -q, --quiet-mode             run commands in quiet mode
```

### SEE ALSO

* [kn](kn.md)	 - kn manages Knative Serving and Eventing resources
* [kn plugin list](kn_plugin_list.md)	 - List plugins

