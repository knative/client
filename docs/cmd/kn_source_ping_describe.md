## kn source ping describe

Show details of a Ping source

### Synopsis

Show details of a Ping source

```
kn source ping describe NAME [flags]
```

### Examples

```

  # Describe a Ping source with name 'myping'
  kn source ping describe myping
```

### Options

```
  -h, --help               help for describe
  -n, --namespace string   Specify the namespace to operate in.
  -v, --verbose            More output.
```

### Options inherited from parent commands

```
      --config string       kn config file (default is ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl config file (default is ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn source ping](kn_source_ping.md)	 - Ping source command group

