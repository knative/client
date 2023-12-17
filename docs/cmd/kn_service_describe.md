## kn service describe

Show details of a service

```
kn service describe NAME
```

### Examples

```

  # Describe service 'svc' in human friendly format
  kn service describe svc

  # Describe service 'svc' in YAML format
  kn service describe svc -o yaml

  # Print only service URL
  kn service describe svc -o url

  # Describe the services in offline mode instead of kubernetes cluster (Beta)
  kn service describe test -n test-ns --target=/user/knfiles
  kn service describe test --target=/user/knfiles/test.yaml
  kn service describe test --target=/user/knfiles/test.json
```

### Options

```
      --allow-missing-template-keys   If true, ignore any errors in templates when a field or map key is missing in the template. Only applies to golang and jsonpath output formats. (default true)
  -h, --help                          help for describe
  -n, --namespace string              Specify the namespace to operate in.
  -o, --output string                 Output format. One of: json|yaml|name|go-template|go-template-file|template|templatefile|jsonpath|jsonpath-as-json|jsonpath-file|url.
      --show-managed-fields           If true, keep the managedFields when printing objects in JSON or YAML format.
      --target string                 Work on local directory instead of a remote cluster (experimental)
      --template string               Template string or path to template file to use when -o=go-template, -o=go-template-file. The template format is golang templates [http://golang.org/pkg/text/template/#pkg-overview].
  -v, --verbose                       More output.
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

