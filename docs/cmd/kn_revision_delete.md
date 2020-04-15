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
      --async              DEPRECATED: please use --no-wait instead. Do not wait for 'revision delete' operation to be completed. (default true)
  -h, --help               help for delete
  -n, --namespace string   Specify the namespace to operate in.
      --no-wait            Do not wait for 'revision delete' operation to be completed. (default true)
      --wait               Wait for 'revision delete' operation to be completed.
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

