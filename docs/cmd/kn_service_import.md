## kn service import

Import a service and its revisions (experimental)

```
kn service import FILENAME
```

### Examples

```

 # Import a service from YAML file
 kn service import /path/to/file.yaml

 # Import a service from JSON file
 kn service import /path/to/file.json
```

### Options

```
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

