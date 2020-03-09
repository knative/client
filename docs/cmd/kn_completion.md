## kn completion

Output shell completion code

### Synopsis


This command prints shell completion code which needs to be evaluated
to provide interactive completion

Supported Shells:
 - bash
 - zsh

```
kn completion [SHELL] [flags]
```

### Examples

```

 # Generate completion code for bash
 source <(kn completion bash)

 # Generate completion code for zsh
 source <(kn completion zsh)
```

### Options

```
  -h, --help   help for completion
```

### Options inherited from parent commands

```
      --config string       kn config file (default is ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl config file (default is ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn](kn.md)	 - Knative client

