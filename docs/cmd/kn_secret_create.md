## kn secret create

Create secret

```
kn secret create NAME
```

### Options

```
  -l, --from-literal strings   Specify comma separated list of key=value pairs.
  -h, --help                   help for create
  -n, --namespace string       Specify the namespace to operate in.
      --tls-cert string        Path to TLS certificate file.
      --tls-key string         Path to TLS key file.
      --type string            Specify Secret type.
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

* [kn secret](kn_secret.md)	 - Manage secrets

