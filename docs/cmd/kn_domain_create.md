## kn domain create

Create a domain mapping

```
kn domain create NAME
```

### Examples

```

  # Create a domain mappings 'hello.example.com' for Knative service 'hello' with TLS secret set
  kn domain create hello.example.com --ref hello --tls my-cert-secret
```

### Options

```
  -h, --help               help for create
  -n, --namespace string   Specify the namespace to operate in.
      --ref string         Addressable target reference for Domain Mapping. You can specify a Knative service, a Knative route. Examples: '--ref' ksvc:hello' or simply '--ref hello' for a Knative service 'hello', '--ref' kroute:hello' for a Knative route 'hello'. '--ref ksvc:mysvc:mynamespace' for a Knative service 'mysvc' in another namespace 'mynamespace', If a prefix is not provided, it is considered as a Knative service in the current namespace. If referring to a Knative service in another namespace, 'ksvc:name:namespace' combination must be provided explicitly.
      --tls string         Enable TLS and point to the secret that holds the server certificate.
```

### Options inherited from parent commands

```
      --cluster string      name of the kubeconfig cluster to use
      --config string       kn configuration file (default: ~/.config/kn/config.yaml)
      --context string      name of the kubeconfig context to use
      --kubeconfig string   kubectl configuration file (default: ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn domain](kn_domain.md)	 - Manage domain mappings

