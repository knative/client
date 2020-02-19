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
      --config string                    kn config file (default is $HOME/.kn/config.yaml)
      --kubeconfig string                kubectl config file (default is $HOME/.kube/config)
      --log-http string[="__STDERR__"]   log http traffic to stderr (no argument) or a file (with argument) (default "__NO_LOG__")
```

### SEE ALSO

* [kn source cronjob](kn_source_cronjob.md)	 - CronJob source command group

