## kn service delete

Delete a service.

### Synopsis

Delete a service.

```
kn service delete NAME [flags]
```

### Examples

```

  # Delete a service 'svc1' in default namespace
  kn service delete svc1

  # Delete a service 'svc2' in 'ns1' namespace
  kn service delete svc2 -n ns1
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

* [kn service](kn_service.md)	 - Service command group

