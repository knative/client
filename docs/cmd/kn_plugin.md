## kn plugin

Plugin command group

### Synopsis

Provides utilities for interacting with kn plugins.

Plugins provide extended functionality that is not part of the major kn command-line distribution.
Please refer to the documentation and examples for more information about how write your own plugins.

```
kn plugin [flags]
```

### Options

```
  -h, --help   help for plugin
```

### Options inherited from parent commands

```
      --config string       config file (default is $HOME/.kn/config.yaml)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --plugin-dir string   kn plugin directory (default is value in kn config or $PATH) (default "$PATH")
```

### SEE ALSO

* [kn](kn.md)	 - Knative client
* [kn plugin list](kn_plugin_list.md)	 - List all visible plugin executables

