## kn completion

Output shell completion code

### Synopsis


This command prints shell completion code which needs to be evaluated
to provide interactive completion

Supported Shells:
 - bash
 - zsh

```
kn completion SHELL
```

### Examples

```

 # Generate completion code for bash
 source <(kn completion bash)

 # Generate completion code for zsh
 source <(kn completion zsh)
 compdef _kn kn
```

### Options

```
  -h, --help   help for completion
```

### Options inherited from parent commands

```
      --as string              username to impersonate for the operation
      --as-group stringArray   group to impersonate for the operation, this flag can be repeated to specify multiple groups
      --cluster string         name of the kubeconfig cluster to use
      --config string          kn configuration file (default: ~/.config/kn/config.yaml)
      --context string         name of the kubeconfig context to use
      --kubeconfig string      kubectl configuration file (default: ~/.kube/config)
      --log-http               log http traffic
```

### SEE ALSO

* [kn](kn.md)	 - kn manages Knative Serving and Eventing resources

