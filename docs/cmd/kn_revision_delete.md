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
  -h, --help               help for delete
  -n, --namespace string   List the requested object(s) in given namespace.
```

### Options inherited from parent commands

```
      --config string       kn config file (default is $HOME/.kn/config.yaml)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn revision](kn_revision.md)	 - Revision command group

