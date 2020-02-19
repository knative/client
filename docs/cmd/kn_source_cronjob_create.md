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
  kn source cronjob create my-cron-trigger --schedule "* * * * */1" --data "ping" --sink svc:mysvc
```

### Options

```
  -d, --data string        String data to send
  -h, --help               help for create
  -n, --namespace string   Specify the namespace to operate in.
      --schedule string    Schedule specification in crontab format (e.g. '* * * * */2' for every two minutes
  -s, --sink string        Addressable sink for events
```

### Options inherited from parent commands

```
      --config string                    kn config file (default is $HOME/.kn/config.yaml)
      --kubeconfig string                kubectl config file (default is $HOME/.kube/config)
      --log-http string[="__STDERR__"]   log http traffic to stderr (no argument) or a file (with argument) (default "__NO_LOG__")
```

### SEE ALSO

* [kn source cronjob](kn_source_cronjob.md)	 - CronJob source command group

