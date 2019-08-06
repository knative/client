## kn plugin list

List all visible plugin executables

### Synopsis

List all visible plugin executables.

Available plugin files are those that are:
- executable
- begin with "kn-
- anywhere on the path specified in Kn's config pluginDir variable, which:
  * can be overridden with the --plugin-dir flag

```
kn plugin list [flags]
```

### Options

```
  -h, --help                     help for list
      --lookup-plugins-in-path   look for kn plugins in $PATH
      --name-only                If true, display only the binary name of each plugin, rather than its full path
      --plugins-dir string       kn plugins directory (default "~/.kn/plugins")
```

### Options inherited from parent commands

```
      --config string       kn config file (default is $HOME/.kn/config.yaml)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn plugin](kn_plugin.md)	 - Plugin command group

