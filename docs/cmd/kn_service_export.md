## kn service export

Export a service and its revisions

### Synopsis

Export a service and its revisions

```
kn service export NAME
```

### Examples

```

  # Export a service in YAML format
  kn service export foo -n bar -o yaml

  # Export a service in JSON format
  kn service export foo -n bar -o json

  # Export a service with revisions
  kn service export foo --with-revisions --mode=export -n bar -o json

  # Export services in kubectl friendly format, as a list kind, one service item for each revision
  kn service export foo --with-revisions --mode=replay -n bar -o json
```

### Options

```
      --allow-missing-template-keys   If true, ignore any errors in templates when a field or map key is missing in the template. Only applies to golang and jsonpath output formats. (default true)
  -h, --help                          help for export
      --mode string                   Format for exporting all routed revisions. One of replay|export (experimental)
  -n, --namespace string              Specify the namespace to operate in.
  -o, --output string                 Output format. One of: json|yaml|name|go-template|go-template-file|template|templatefile|jsonpath|jsonpath-file.
      --template string               Template string or path to template file to use when -o=go-template, -o=go-template-file. The template format is golang templates [http://golang.org/pkg/text/template/#pkg-overview].
      --with-revisions                Export all routed revisions (experimental)
```

### Options inherited from parent commands

```
      --config string       kn configuration file (default: ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl configuration file (default: ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn service](kn_service.md)	 - Manage Knative services

