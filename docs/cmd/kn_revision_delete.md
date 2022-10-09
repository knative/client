## kn revision delete

Delete revisions

```
kn revision delete NAME [NAME ...]
```

### Examples

```

  # Delete a revision 'svc1-abcde' in default namespace
  kn revision delete svc1-abcde

  # Delete all unreferenced revisions
  kn revision delete --prune-all

  # Delete all unreferenced revisions for a given service 'mysvc'
  kn revision delete --prune mysvc
```

### Options

```
  -h, --help               help for delete
  -n, --namespace string   Specify the namespace to operate in.
      --no-wait            Do not wait for 'revision delete' operation to be completed. (default true)
      --prune string       Remove unreferenced revisions for a given service in a namespace.
      --prune-all          Remove all unreferenced revisions in a namespace.
      --wait               Wait for 'revision delete' operation to be completed.
      --wait-timeout int   Seconds to wait before giving up on waiting for revision to be deleted. (default 600)
      --wait-window int    Seconds to wait for revision to be deleted after a false ready condition is returned (default 2)
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

* [kn revision](kn_revision.md)	 - Manage service revisions

