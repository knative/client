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

  # Delete all services in 'ns1' namespace
  kn service delete --all -n ns1
```

### Options

```
      --all                Delete all services in a namespace.
      --async              DEPRECATED: please use --no-wait instead. Do not wait for 'service delete' operation to be completed. (default true)
  -h, --help               help for delete
  -n, --namespace string   Specify the namespace to operate in.
      --no-wait            Do not wait for 'service delete' operation to be completed. (default true)
      --wait               Wait for 'service delete' operation to be completed.
      --wait-timeout int   Seconds to wait before giving up on waiting for service to be deleted. (default 600)
```

### Options inherited from parent commands

```
      --config string       kn configuration file (default: ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl configuration file (default: ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn service](kn_service.md)	 - Service command group

