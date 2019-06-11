## kn plugin list

List all visible plugin executables on a user's PATH

### Synopsis

List all visible plugin executables on a user's PATH.

		Available plugin files are those that are:
		- executable
		- anywhere on the user's PATH
		- begin with "kn-

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
      --plugin-dir string   kn plugin directory (default is value in kn config or $PATH)
```

### SEE ALSO

* [kn plugin](kn_plugin.md)	 - Plugin command group

