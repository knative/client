## kn source cronjob describe

Show details of a CronJob source

### Synopsis

Show details of a CronJob source

```
kn source cronjob describe NAME [flags]
```

### Examples

```

  # Describe a cronjob source with name 'mycronjob'
  kn source cronjob describe mycronjob
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

* [kn source cronjob](kn_source_cronjob.md)	 - CronJob source command group

