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
      --async              DEPRECATED: please use --no-wait instead. Delete service and don't wait for it to be deleted.
  -h, --help               help for delete
  -n, --namespace string   Specify the namespace to operate in.
      --no-wait            Delete service and don't wait for it to be deleted.
      --wait-timeout int   Seconds to wait before giving up on waiting for service to be deleted. (default 600)
```

### Options inherited from parent commands

```
      --config string       kn config file (default is ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl config file (default is ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn service](kn_service.md)	 - Service command group

