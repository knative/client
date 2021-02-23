## kn service update

Update a service

```
kn service update NAME
```

### Examples

```

  # Updates a service 'svc' with new environment variables
  kn service update svc --env KEY1=VALUE1 --env KEY2=VALUE2

  # Update a service 'svc' with new port
  kn service update svc --port 80

  # Updates a service 'svc' with new request and limit parameters
  kn service update svc --request cpu=500m --limit memory=1024Mi --limit cpu=1000m

  # Assign tag 'latest' and 'stable' to revisions 'echo-v2' and 'echo-v1' respectively
  kn service update svc --tag echo-v2=latest --tag echo-v1=stable
  OR
  kn service update svc --tag echo-v2=latest,echo-v1=stable

  # Update tag from 'testing' to 'staging' for latest ready revision of service
  kn service update svc --untag testing --tag @latest=staging

  # Add tag 'test' to echo-v3 revision with 10% traffic and rest to latest ready revision of service
  kn service update svc --tag echo-v3=test --traffic test=10,@latest=90

  # Update the service in offline mode instead of kubernetes cluster
  kn service update gitopstest -n test-ns --env KEY1=VALUE1 --target=/user/knfiles
  kn service update gitopstest --env KEY1=VALUE1 --target=/user/knfiles/test.yaml
  kn service update gitopstest --env KEY1=VALUE1 --target=/user/knfiles/test.json
```

### Options

```
  -a, --annotation stringArray            Annotations to set for both Service and Revision. name=value; you may provide this flag any number of times to set multiple annotations. To unset, specify the annotation name followed by a "-" (e.g., name-).
      --annotation-revision stringArray   Revision annotation to set. name=value; you may provide this flag any number of times to set multiple annotations. To unset, specify the annotation name followed by a "-" (e.g., name-). This flag takes precedence over the "annotation" flag.
      --annotation-service stringArray    Service annotation to set. name=value; you may provide this flag any number of times to set multiple annotations. To unset, specify the annotation name followed by a "-" (e.g., name-). This flag takes precedence over the "annotation" flag.
      --arg stringArray                   Add argument to the container command. Example: --arg myArg1 --arg --myArg2 --arg myArg3=3. You can use this flag multiple times.
      --autoscale-window string           Duration to look back for making auto-scaling decisions. The service is scaled to zero if no request was received in during that time. (eg: 10s)
      --cluster-local                     Specify that the service be private. (--no-cluster-local will make the service publicly available)
      --cmd string                        Specify command to be used as entrypoint instead of default one. Example: --cmd /app/start or --cmd /app/start --arg myArg to pass additional arguments.
      --concurrency-limit int             Hard Limit of concurrent requests to be processed by a single replica.
      --concurrency-target int            Recommendation for when to scale up based on the concurrent number of incoming request. Defaults to --concurrency-limit when given.
      --concurrency-utilization int       Percentage of concurrent requests utilization before scaling up. (default 70)
  -e, --env stringArray                   Environment variable to set. NAME=value; you may provide this flag any number of times to set multiple environment variables. To unset, specify the environment variable name followed by a "-" (e.g., NAME-).
      --env-from stringArray              Add environment variables from a ConfigMap (prefix cm: or config-map:) or a Secret (prefix secret:). Example: --env-from cm:myconfigmap or --env-from secret:mysecret. You can use this flag multiple times. To unset a ConfigMap/Secret reference, append "-" to the name, e.g. --env-from cm:myconfigmap-.
  -h, --help                              help for update
      --image string                      Image to run.
  -l, --label stringArray                 Labels to set for both Service and Revision. name=value; you may provide this flag any number of times to set multiple labels. To unset, specify the label name followed by a "-" (e.g., name-).
      --label-revision stringArray        Revision label to set. name=value; you may provide this flag any number of times to set multiple labels. To unset, specify the label name followed by a "-" (e.g., name-). This flag takes precedence over the "label" flag.
      --label-service stringArray         Service label to set. name=value; you may provide this flag any number of times to set multiple labels. To unset, specify the label name followed by a "-" (e.g., name-). This flag takes precedence over the "label" flag.
      --limit strings                     The resource requirement limits for this Service. For example, 'cpu=100m,memory=256Mi'. You can use this flag multiple times. To unset a resource limit, append "-" to the resource name, e.g. '--limit memory-'.
      --lock-to-digest                    Keep the running image for the service constant when not explicitly specifying the image. (--no-lock-to-digest pulls the image tag afresh with each new revision) (default true)
      --mount stringArray                 Mount a ConfigMap (prefix cm: or config-map:), a Secret (prefix secret: or sc:), or an existing Volume (without any prefix) on the specified directory. Example: --mount /mydir=cm:myconfigmap, --mount /mydir=secret:mysecret, or --mount /mydir=myvolume. When a configmap or a secret is specified, a corresponding volume is automatically generated. You can use this flag multiple times. For unmounting a directory, append "-", e.g. --mount /mydir-, which also removes any auto-generated volume.
  -n, --namespace string                  Specify the namespace to operate in.
      --no-cluster-local                  Do not specify that the service be private. (--no-cluster-local will make the service publicly available) (default true)
      --no-lock-to-digest                 Do not keep the running image for the service constant when not explicitly specifying the image. (--no-lock-to-digest pulls the image tag afresh with each new revision)
      --no-wait                           Do not wait for 'service update' operation to be completed.
  -p, --port string                       The port where application listens on, in the format 'NAME:PORT', where 'NAME' is optional. Examples: '--port h2c:8080' , '--port 8080'.
      --pull-secret string                Image pull secret to set. An empty argument ("") clears the pull secret. The referenced secret must exist in the service's namespace.
      --request strings                   The resource requirement requests for this Service. For example, 'cpu=100m,memory=256Mi'. You can use this flag multiple times. To unset a resource request, append "-" to the resource name, e.g. '--request cpu-'.
      --revision-name string              The revision name to set. Must start with the service name and a dash as a prefix. Empty revision name will result in the server generating a name for the revision. Accepts golang templates, allowing {{.Service}} for the service name, {{.Generation}} for the generation, and {{.Random [n]}} for n random consonants. (default "{{.Service}}-{{.Random 5}}-{{.Generation}}")
      --scale string                      Set the Minimum and Maximum number of replicas. You can use this flag to set both to a single value, or set a range with min/max values, or set either min or max values without specifying the other. Example: --scale 5 (scale-min = 5, scale-max = 5) or --scale 1..5 (scale-min = 1, scale-max = 5) or --scale 1.. (scale-min = 1, scale-max = unchanged) or --scale ..5 (scale-min = unchanged, scale-max = 5)
      --scale-init int                    Initial number of replicas with which a service starts. Can be 0 or a positive integer.
      --scale-max int                     Maximum number of replicas.
      --scale-min int                     Minimum number of replicas.
      --service-account string            Service account name to set. An empty argument ("") clears the service account. The referenced service account must exist in the service's namespace.
      --tag strings                       Set tag (format: --tag revisionRef=tagName) where revisionRef can be a revision or '@latest' string representing latest ready revision. This flag can be specified multiple times.
      --target string                     Work on local directory instead of a remote cluster (experimental)
      --traffic strings                   Set traffic distribution (format: --traffic revisionRef=percent) where revisionRef can be a revision or a tag or '@latest' string representing latest ready revision. This flag can be given multiple times with percent summing up to 100%.
      --untag strings                     Untag revision (format: --untag tagName). This flag can be specified multiple times.
      --user int                          The user ID to run the container (e.g., 1001).
      --volume stringArray                Add a volume from a ConfigMap (prefix cm: or config-map:) or a Secret (prefix secret: or sc:). Example: --volume myvolume=cm:myconfigmap or --volume myvolume=secret:mysecret. You can use this flag multiple times. To unset a ConfigMap/Secret reference, append "-" to the name, e.g. --volume myvolume-.
      --wait                              Wait for 'service update' operation to be completed. (default true)
      --wait-timeout int                  Seconds to wait before giving up on waiting for service to be ready. (default 600)
```

### Options inherited from parent commands

```
      --cluster string      name of the kubeconfig cluster to use
      --config string       kn configuration file (default: ~/.config/kn/config.yaml)
      --context string      name of the kubeconfig context to use
      --kubeconfig string   kubectl configuration file (default: ~/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn service](kn_service.md)	 - Manage Knative services

