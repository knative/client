## kn service import

Import a service and its revisions

### Synopsis

Import a service and its revisions

```
kn service import FILENAME
```

### Examples

```

 # Import a service in YAML format
 kn service import /path/to/file.yaml

 # Import a service in JSON format
 kn service import /path/to/file.json
```

### Options

```
      --async              DEPRECATED: please use --no-wait instead. Do not wait for 'service import' operation to be completed.
  -h, --help               help for import
  -n, --namespace string   Specify the namespace to operate in.
      --no-wait            Do not wait for 'service import' operation to be completed.
      --wait               Wait for 'service import' operation to be completed. (default true)
      --wait-timeout int   Seconds to wait before giving up on waiting for service to be ready. (default 600)
```

### Options inherited from parent commands

```
      --config string       kn configuration file (default: ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl configuration file (default: ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn service](kn_service.md)	 - Manage Knative services

