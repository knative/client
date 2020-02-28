## kn plugin

Plugin command group

### Synopsis

Provides utilities for interacting and managing with kn plugins.

Plugins provide extended functionality that is not part of the core kn command-line distribution.
Please refer to the documentation and examples for more information about how write your own plugins.

```
kn plugin [flags]
```

### Options

```
  -h, --help                 help for plugin
      --lookup-plugins       look for kn plugins in $PATH
      --plugins-dir string   kn plugins directory (default "~/.config/kn/plugins")
```

### Options inherited from parent commands

```
      --config string       kn config file (default is ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl config file (default is ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn](kn.md)	 - Knative client
* [kn plugin list](kn_plugin_list.md)	 - List plugins

