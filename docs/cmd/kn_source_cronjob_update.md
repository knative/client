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
  kn source cronjob update my-cron-trigger --schedule "* * * * *"
```

### Options

```
  -d, --data string              String data to send
  -h, --help                     help for update
      --limits-cpu string        The limits on the requested CPU (e.g., 1000m).
      --limits-memory string     The limits on the requested memory (e.g., 1024Mi).
  -n, --namespace string         Specify the namespace to operate in.
      --requests-cpu string      The requested CPU (e.g., 250m).
      --requests-memory string   The requested memory (e.g., 64Mi).
      --schedule string          Schedule specification in crontab format (e.g. '*/2 * * * *' for every two minutes
      --service-account string   Name of the service account to use to run this source
  -s, --sink string              Addressable sink for events
```

### Options inherited from parent commands

```
      --config string       kn config file (default is ~/.config/kn/config.yaml)
      --kubeconfig string   kubectl config file (default is ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn source cronjob](kn_source_cronjob.md)	 - CronJob source command group

