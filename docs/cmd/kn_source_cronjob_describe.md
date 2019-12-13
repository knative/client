## kn source cronjob describe

Describe a CronJob source.

### Synopsis

Describe a CronJob source.

```
kn source cronjob describe NAME [flags]
```

### Examples

```

  # Describe a cronjob source with name 'my-cron-trigger'
  kn source cronjob describe my-cron-trigger
```

### Options

```
  -h, --help               help for describe
  -n, --namespace string   Specify the namespace to operate in.
  -v, --verbose            More output.
```

### Options inherited from parent commands

```
      --config string       kn config file (default is $HOME/.kn/config.yaml)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn source cronjob](kn_source_cronjob.md)	 - CronJob source command group

