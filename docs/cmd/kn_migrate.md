## kn migrate

Migrate knative services from source cluster to destination cluster

### Synopsis

Migrate knative services from source cluster to destination cluster

```
kn migrate [flags]
```

### Options

```
      --delete                          Delete all Knative resources after kn-migration from source cluster
      --destination-kubeconfig string   The kubeconfig of the destination Knative resources (default is KUBECONFIG2 from ENV property)
      --destination-namespace string    The namespace of the destination Knative resources (default "default")
      --force                           Migrate service forcefully, replaces existing service if any.
  -h, --help                            help for migrate
  -n, --namespace string                The namespace of the source Knative resources (default "default")
```

### Options inherited from parent commands

```
      --config string       kn config file (default is $HOME/.kn/config.yaml)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn](kn.md)	 - Knative client

