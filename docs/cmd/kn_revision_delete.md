## kn revision delete

Delete revisions

```
kn revision delete NAME [NAME ...]
```

### Examples

```

  # Delete a revision 'svc1-abcde' in default namespace
  kn revision delete svc1-abcde
```

### Options

```
  -h, --help               help for delete
  -n, --namespace string   Specify the namespace to operate in.
      --no-wait            Do not wait for 'revision delete' operation to be completed. (default true)
      --wait               Wait for 'revision delete' operation to be completed.
      --wait-timeout int   Seconds to wait before giving up on waiting for revision to be deleted. (default 600)
```

### Options inherited from parent commands

```
      --config string       kn configuration file (default: ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl configuration file (default: ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn revision](kn_revision.md)	 - Manage service revisions

