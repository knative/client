## kn service export

Export a service.

### Synopsis

Export a service.

```
kn service export NAME [flags]
```

### Examples

```

  # Export a service in YAML format
  kn service export foo -n bar -o yaml
  # Export a service in JSON format
  kn service export foo -n bar -o json
  # Export a service with revisions
  kn service export foo --with-revisions --mode=resources -n bar -o json
  # Export services in kubectl friendly format, as a list kind, one service item for each revision
  kn service export foo --with-revisions --mode=kubernetes -n bar -o json
```

### Options

```
      --allow-missing-template-keys   If true, ignore any errors in templates when a field or map key is missing in the template. Only applies to golang and jsonpath output formats. (default true)
  -h, --help                          help for export
      --mode string                   Format for exporting all routed revisions. One of kubernetes|resources (experimental)
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

* [kn service](kn_service.md)	 - Service command group

