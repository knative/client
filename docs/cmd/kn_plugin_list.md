## kn plugin list

List plugins

### Synopsis

List all installed plugins.

Available plugins are those that are:
- executable
- begin with "kn-"
- Kn's plugin directory
- Anywhere in the execution $PATH (if plugins.path-lookup config variable is enabled)

```
kn plugin list [flags]
```

### Options

```
  -h, --help      help for list
      --verbose   verbose output
```

### Options inherited from parent commands

```
      --config string       kn configuration file (default: ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl configuration file (default: ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn plugin](kn_plugin.md)	 - Plugin command group

