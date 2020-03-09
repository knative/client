## kn source ping list

List Ping sources.

### Synopsis

List Ping sources.

```
kn source ping list [flags]
```

### Examples

```

  # List all Ping sources
  kn source ping list

  # List all Ping sources in YAML format
  kn source ping list -o yaml
```

### Options

```
  -A, --all-namespaces                If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.
      --allow-missing-template-keys   If true, ignore any errors in templates when a field or map key is missing in the template. Only applies to golang and jsonpath output formats. (default true)
  -h, --help                          help for list
  -n, --namespace string              Specify the namespace to operate in.
      --no-headers                    When using the default output format, don't print headers (default: print headers).
  -o, --output string                 Output format. One of: json|yaml|name|go-template|go-template-file|template|templatefile|jsonpath|jsonpath-file.
      --template string               Template string or path to template file to use when -o=go-template, -o=go-template-file. The template format is golang templates [http://golang.org/pkg/text/template/#pkg-overview].
```

### Options inherited from parent commands

```
      --config string       kn config file (default is ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl config file (default is ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn source ping](kn_source_ping.md)	 - Ping source command group

