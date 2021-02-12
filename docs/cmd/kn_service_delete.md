## kn service delete

Delete services

```
kn service delete NAME [NAME ...]
```

### Examples

```

  # Delete a service 'svc1' in default namespace
  kn service delete svc1

  # Delete a service 'svc2' in 'ns1' namespace
  kn service delete svc2 -n ns1

  # Delete all services in 'ns1' namespace
  kn service delete --all -n ns1

  # Delete the services in offline mode instead of kubernetes cluster
  kn service delete test -n test-ns --target=/user/knfiles
  kn service delete test --target=/user/knfiles/test.yaml
  kn service delete test --target=/user/knfiles/test.json
```

### Options

```
      --all                Delete all services in a namespace.
  -h, --help               help for delete
  -n, --namespace string   Specify the namespace to operate in.
      --no-wait            Do not wait for 'service delete' operation to be completed. (default true)
      --target string      Work on local directory instead of a remote cluster (experimental)
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

* [kn service](kn_service.md)	 - Manage Knative services

