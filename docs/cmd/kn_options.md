## kn options

Print the list of flags inherited by all commands

### Synopsis

Print the list of flags inherited by all commands

```
kn options [flags]
```

### Examples

```
# Print flags inherited by all commands
kn options
```

### Options

```
  -h, --help   help for options
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
  -q, --quiet                  run commands in quiet mode
```

### SEE ALSO

* [kn](kn.md)	 - kn manages Knative Serving and Eventing resources

