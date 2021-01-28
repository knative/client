## kn source binding describe

Show details of a sink binding

### Synopsis

Show details of a sink binding

```
kn source binding describe NAME
```

### Examples

```

  # Describe a sink binding 'mysinkbinding'
  kn source binding describe mysinkbinding

  # Describe a sink binding 'mysinkbinding' in YAML format
  kn source binding describe mysinkbinding -o yaml
```

### Options

```
      --allow-missing-template-keys   If true, ignore any errors in templates when a field or map key is missing in the template. Only applies to golang and jsonpath output formats. (default true)
  -h, --help                          help for describe
  -n, --namespace string              Specify the namespace to operate in.
  -o, --output string                 Output format. One of: json|yaml|name|go-template|go-template-file|template|templatefile|jsonpath|jsonpath-as-json|jsonpath-file.
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

* [kn source binding](kn_source_binding.md)	 - Manage sink bindings

