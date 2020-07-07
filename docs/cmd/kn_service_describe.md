## kn service describe

Show details of a service

### Synopsis

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
```

### Options

```
      --allow-missing-template-keys   If true, ignore any errors in templates when a field or map key is missing in the template. Only applies to golang and jsonpath output formats. (default true)
  -h, --help                          help for describe
  -n, --namespace string              Specify the namespace to operate in.
  -o, --output string                 Output format one of json|yaml|name|go-template|go-template-file|template|templatefile|jsonpath|jsonpath-file|url.
      --template string               Template string or path to template file to use when -o=go-template, -o=go-template-file. The template format is golang templates [http://golang.org/pkg/text/template/#pkg-overview].
  -v, --verbose                       More output.
```

### Options inherited from parent commands

```
      --config string       kn configuration file (default: ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl configuration file (default: ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn service](kn_service.md)	 - Manage Knative services

