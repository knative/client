## kn service import

Import a service and its revisions (experimental)

```
kn service import FILENAME
```

### Examples

```

 # Import a service from YAML file (Beta)
 kn service import /path/to/file.yaml

 # Import a service from JSON file (Beta)
 kn service import /path/to/file.json
```

### Options

```
  -h, --help               help for import
  -n, --namespace string   Specify the namespace to operate in.
      --no-wait            Do not wait for 'service import' operation to be completed.
      --wait               Wait for 'service import' operation to be completed. (default true)
      --wait-timeout int   Seconds to wait before giving up on waiting for service to be ready. (default 600)
      --wait-window int    Seconds to wait for service to be ready after a false ready condition is returned (default 2)
```

### Options inherited from parent commands

```
      --as string              username to impersonate for the operation
      --as-group stringArray   group to impersonate for the operation, this flag can be repeated to specify multiple groups
      --as-uid string          uid to impersonate for the operation
      --cluster string         name of the kubeconfig cluster to use
      --config string          kn configuration file (default: ~/.config/kn/config.yaml)
      --context string         name of the kubeconfig context to use
      --kubeconfig string      kubectl configuration file (default: ~/.kube/config)
      --log-http               log http traffic
  -q, --quiet                  run commands in quiet mode
```

### SEE ALSO

* [kn service](kn_service.md)	 - Manage Knative services

