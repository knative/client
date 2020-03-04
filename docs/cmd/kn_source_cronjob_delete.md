## kn source cronjob delete

Delete a CronJob source.

### Synopsis

Delete a CronJob source.

```
kn source cronjob delete NAME [flags]
```

### Examples

```

  # Delete a CronJob source 'my-cron-trigger'
  kn source cronjob delete my-cron-trigger
```

### Options

```
  -h, --help               help for delete
  -n, --namespace string   Specify the namespace to operate in.
```

### Options inherited from parent commands

```
      --config string       kn config file (default is ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl config file (default is ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn source cronjob](kn_source_cronjob.md)	 - CronJob source command group

