## kn broker describe

Describe broker

### Synopsis

Describe broker

```
kn broker describe NAME
```

### Examples

```

  # Describe broker 'mybroker' in the current namespace
  kn broker describe mybroker

  # Describe broker 'mybroker' in the 'myproject' namespace
  kn broker describe mybroker --namespace myproject

  # Describe broker 'mybroker' in YAML format
  kn broker describe mybroker -o yaml

  # Print only broker URL
  kn broker describe mybroker -o url
```

### Options

```
      --allow-missing-template-keys   If true, ignore any errors in templates when a field or map key is missing in the template. Only applies to golang and jsonpath output formats. (default true)
  -h, --help                          help for describe
  -n, --namespace string              Specify the namespace to operate in.
  -o, --output string                 Output format. One of: json|yaml|name|go-template|go-template-file|template|templatefile|jsonpath|jsonpath-as-json|jsonpath-file|url.
      --template string               Template string or path to template file to use when -o=go-template, -o=go-template-file. The template format is golang templates [http://golang.org/pkg/text/template/#pkg-overview].
```

### Options inherited from parent commands

```
      --config string       kn configuration file (default: ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl configuration file (default: ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn broker](kn_broker.md)	 - Manage message brokers

