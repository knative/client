## kn service export

Export a service and its revisions

```
kn service export NAME
```

### Examples

```

  # Export a service in YAML format (Beta)
  kn service export foo -n bar -o yaml

  # Export a service in JSON format (Beta)
  kn service export foo -n bar -o json

  # Export a service with revisions (Beta)
  kn service export foo --with-revisions --mode=export -n bar -o json

  # Export services in kubectl friendly format, as a list kind, one service item for each revision (Beta)
  kn service export foo --with-revisions --mode=replay -n bar -o json
```

### Options

```
      --allow-missing-template-keys   If true, ignore any errors in templates when a field or map key is missing in the template. Only applies to golang and jsonpath output formats. (default true)
  -h, --help                          help for export
      --mode string                   Format for exporting all routed revisions. One of replay|export (Beta)
  -n, --namespace string              Specify the namespace to operate in.
  -o, --output string                 Output format. One of: (json, yaml, name, go-template, go-template-file, template, templatefile, jsonpath, jsonpath-as-json, jsonpath-file).
      --show-managed-fields           If true, keep the managedFields when printing objects in JSON or YAML format.
      --template string               Template string or path to template file to use when -o=go-template, -o=go-template-file. The template format is golang templates [http://golang.org/pkg/text/template/#pkg-overview].
      --with-revisions                Export all routed revisions (Beta)
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
```

### SEE ALSO

* [kn service](kn_service.md)	 - Manage Knative services

