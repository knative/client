## kn importer cron create

Create an Crontab scheduler as importer

### Synopsis

Create an Crontab scheduler as importer

```
kn importer cron create NAME --crontab SCHEDULE --sink SERVICE --data DATA [flags]
```

### Examples

```

  # Create a crontabs scheduler 'mycron' which fires every minute and sends 'ping'' to service 'mysvc' as a cloudevent
  kn importer cron create mycron --schedule "* * * * */1" --data "ping" --sink svc:mysvc
```

### Options

```
  -d, --data string        Data to send
  -h, --help               help for create
  -n, --namespace string   List the requested object(s) in given namespace.
      --schedule string    Schedule specification in crontab format
  -s, --sink string        Addressable sink for events
```

### Options inherited from parent commands

```
      --config string       kn config file (default is $HOME/.kn/config.yaml)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn importer cron](kn_importer_cron.md)	 - Cron source command group

