## kn source cronjob create

Create a CronJob source.

### Synopsis

Create a CronJob source.

```
kn source cronjob create NAME --schedule SCHEDULE --sink SINK --data DATA [flags]
```

### Examples

```

  # Create a crontab scheduler 'my-cron-trigger' which fires every minute and sends 'ping' to service 'mysvc' as a cloudevent
  kn source cronjob create my-cron-trigger --schedule "* * * * *" --data "ping" --sink svc:mysvc
  
  # Create a crontab scheduler 'my-cron-trigger' with ServiceAccount name
  kn source cronjob create my-cron-trigger --schedule "* * * * *" --data "ping" --sink svc:event-display --service-account myaccount

  # Create a crontab scheduler 'my-cron-trigger' with requested resources
  kn source cronjob create my-cron-trigger --schedule "* * * * *" --data "ping" --sink svc:event-display --requests-cpu 100m --requests-memory 128Mi
```

### Options

```
  -d, --data string              String data to send
  -h, --help                     help for create
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

