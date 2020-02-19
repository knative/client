## kn source apiserver list

List ApiServer sources.

### Synopsis

List ApiServer sources.

```
kn source apiserver list [flags]
```

### Examples

```

  # List all ApiServer sources in YAML format
  kn source apiserver list -o yaml
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
      --config string                    kn config file (default is $HOME/.kn/config.yaml)
      --kubeconfig string                kubectl config file (default is $HOME/.kube/config)
      --log-http string[="__STDERR__"]   log http traffic to stderr (no argument) or a file (with argument) (default "__NO_LOG__")
```

### SEE ALSO

* [kn source apiserver](kn_source_apiserver.md)	 - Kubernetes API Server Event Source command group

