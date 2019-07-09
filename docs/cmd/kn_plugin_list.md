## kn plugin list

List all visible plugin executables

### Synopsis

List all visible plugin executables.

		Available plugin files are those that are:
		- executable
		- begin with "kn-
		- anywhere on the path specfied in Kn's config pluginDir variable, which:
		  * defaults to $PATH if not specified
		  * can be overridden with the --plugin-dir flag

```
kn plugin list [flags]
```

### Options

```
  -h, --help        help for list
      --name-only   If true, display only the binary name of each plugin, rather than its full path
```

### Options inherited from parent commands

```
      --config string       config file (default is $HOME/.kn/config.yaml)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --plugin-dir string   kn plugin directory (default is value in kn config or $PATH) (default "$PATH")
```

### SEE ALSO

* [kn plugin](kn_plugin.md)	 - Plugin command group

