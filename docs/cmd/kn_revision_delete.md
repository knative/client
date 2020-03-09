## kn revision delete

Delete a revision.

### Synopsis

Delete a revision.

```
kn revision delete NAME [flags]
```

### Examples

```

  # Delete a revision 'svc1-abcde' in default namespace
  kn revision delete svc1-abcde
```

### Options

```
      --async              DEPRECATED: please use --no-wait instead. Delete revision and don't wait for it to be deleted.
  -h, --help               help for delete
  -n, --namespace string   Specify the namespace to operate in.
      --no-wait            Delete revision and don't wait for it to be deleted.
      --wait-timeout int   Seconds to wait before giving up on waiting for revision to be deleted. (default 600)
```

### Options inherited from parent commands

```
      --config string       kn config file (default is ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl config file (default is ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn revision](kn_revision.md)	 - Revision command group

