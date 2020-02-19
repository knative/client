## kn route describe

Show details of a route

### Synopsis

Show details of a route

```
kn route describe NAME [flags]
```

### Options

```
      --allow-missing-template-keys   If true, ignore any errors in templates when a field or map key is missing in the template. Only applies to golang and jsonpath output formats. (default true)
  -h, --help                          help for describe
  -n, --namespace string              Specify the namespace to operate in.
  -o, --output string                 Output format. One of: json|yaml|name|go-template|go-template-file|template|templatefile|jsonpath|jsonpath-file.
      --template string               Template string or path to template file to use when -o=go-template, -o=go-template-file. The template format is golang templates [http://golang.org/pkg/text/template/#pkg-overview].
  -v, --verbose                       More output.
```

### Options inherited from parent commands

```
      --config string                    kn config file (default is $HOME/.kn/config.yaml)
      --kubeconfig string                kubectl config file (default is $HOME/.kube/config)
      --log-http string[="__STDERR__"]   log http traffic to stderr (no argument) or a file (with argument) (default "__NO_LOG__")
```

### SEE ALSO

* [kn route](kn_route.md)	 - Route command group

