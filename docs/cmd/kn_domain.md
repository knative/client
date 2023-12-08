## kn domain

Manage domain mappings

```
kn domain COMMAND
```

### Options

```
  -h, --help   help for domain
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
  -q, --quiet-mode             run commands in quiet mode
```

### SEE ALSO

* [kn](kn.md)	 - kn manages Knative Serving and Eventing resources
* [kn domain create](kn_domain_create.md)	 - Create a domain mapping
* [kn domain delete](kn_domain_delete.md)	 - Delete a domain mapping
* [kn domain describe](kn_domain_describe.md)	 - Show details of a domain mapping
* [kn domain list](kn_domain_list.md)	 - List domain mappings
* [kn domain update](kn_domain_update.md)	 - Update a domain mapping

