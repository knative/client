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
      --plugins-dir string   kn plugins directory (default "~/.kn/plugins")
```

### Options inherited from parent commands

```
      --config string                    kn config file (default is $HOME/.kn/config.yaml)
      --kubeconfig string                kubectl config file (default is $HOME/.kube/config)
      --log-http string[="__STDERR__"]   log http traffic to stderr (no argument) or a file (with argument) (default "__NO_LOG__")
```

### SEE ALSO

* [kn](kn.md)	 - Knative client
* [kn plugin list](kn_plugin_list.md)	 - List plugins

