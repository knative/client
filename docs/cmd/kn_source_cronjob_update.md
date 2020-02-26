## kn source cronjob update

Update a CronJob source.

### Synopsis

Update a CronJob source.

```
kn source cronjob update NAME --schedule SCHEDULE --sink SERVICE --data DATA [flags]
```

### Examples

```

  # Update the schedule of a crontab source 'my-cron-trigger' to fire every minute
  kn source cronjob update my-cron-trigger --schedule "* * * * */1"
```

### Options

```
  -d, --data string        String data to send
  -h, --help               help for update
  -n, --namespace string   Specify the namespace to operate in.
      --schedule string    Schedule specification in crontab format (e.g. '* * * * */2' for every two minutes
  -s, --sink string        Addressable sink for events
```

### Options inherited from parent commands

```
      --config string       kn config file (default is ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl config file (default is ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn source cronjob](kn_source_cronjob.md)	 - CronJob source command group

