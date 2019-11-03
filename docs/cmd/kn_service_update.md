## kn service update

Update a service.

### Synopsis

Update a service.

```
kn service update NAME [flags]
```

### Examples

```

  # Updates a service 'svc' with new environment variables
  kn service update svc --env KEY1=VALUE1 --env KEY2=VALUE2

  # Update a service 'svc' with new port
  kn service update svc --port 80

  # Updates a service 'svc' with new requests and limits parameters
  kn service update svc --requests-cpu 500m --limits-memory 1024Mi

  # Assign tag 'latest' and 'stable' to revisions 'echo-v2' and 'echo-v1' respectively
  kn service update svc --tag echo-v2=latest --tag echo-v1=stable
  OR
  kn service update svc --tag echo-v2=latest,echo-v1=stable

  # Update tag from 'testing' to 'staging' for latest ready revision of service
  kn service update svc --untag testing --tag @latest=staging

  # Add tag 'test' to echo-v3 revision with 10% traffic and rest to latest ready revision of service
  kn service update svc --tag echo-v3=test --traffic test=10,@latest=90
```

### Options

```
      --annotation stringArray     Service annotation to set. name=value; you may provide this flag any number of times to set multiple annotations. To unset, specify the annotation name followed by a "-" (e.g., name-).
      --async                      Update service and don't wait for it to become ready.
      --concurrency-limit int      Hard Limit of concurrent requests to be processed by a single replica.
      --concurrency-target int     Recommendation for when to scale up based on the concurrent number of incoming request. Defaults to --concurrency-limit when given.
  -e, --env stringArray            Environment variable to set. NAME=value; you may provide this flag any number of times to set multiple environment variables. To unset, specify the environment variable name followed by a "-" (e.g., NAME-).
      --env-from stringArray       Config a envfrom with a config map or secret. (config-map | secret):CONFIG_MAP_OR_SECRET_NAME you may provide this flag any number of times. To unset, specify the config map name followed by a "-" (e.g., config-map:CONFIG_MAP_NAME- or secret:SECRET_NAME-).
  -h, --help                       help for update
      --image string               Image to run.
  -l, --label stringArray          Service label to set. name=value; you may provide this flag any number of times to set multiple labels. To unset, specify the label name followed by a "-" (e.g., name-).
      --limits-cpu string          The limits on the requested CPU (e.g., 1000m).
      --limits-memory string       The limits on the requested memory (e.g., 1024Mi).
      --lock-to-digest             keep the running image for the service constant when not explicitly specifying the image. (--no-lock-to-digest pulls the image tag afresh with each new revision) (default true)
      --max-scale int              Maximal number of replicas.
      --min-scale int              Minimal number of replicas.
  -n, --namespace string           Specify the namespace to operate in.
      --no-lock-to-digest          do not keep the running image for the service constant when not explicitly specifying the image. (--no-lock-to-digest pulls the image tag afresh with each new revision)
  -p, --port int32                 The port where application listens on.
      --requests-cpu string        The requested CPU (e.g., 250m).
      --requests-memory string     The requested memory (e.g., 64Mi).
      --revision-name string       The revision name to set. Must start with the service name and a dash as a prefix. Empty revision name will result in the server generating a name for the revision. Accepts golang templates, allowing {{.Service}} for the service name, {{.Generation}} for the generation, and {{.Random [n]}} for n random consonants. (default "{{.Service}}-{{.Random 5}}-{{.Generation}}")
      --service-account string     Service account name to set. Empty service account name will result to clear the service account.
      --tag strings                Set tag (format: --tag revisionRef=tagName) where revisionRef can be a revision or '@latest' string representing latest ready revision. This flag can be specified multiple times.
      --traffic strings            Set traffic distribution (format: --traffic revisionRef=percent) where revisionRef can be a revision or a tag or '@latest' string representing latest ready revision. This flag can be given multiple times with percent summing up to 100%.
      --untag strings              Untag revision (format: --untag tagName). This flag can be specified multiple times.
      --volume stringArray         Config a volume with a config map or secret. VOLUME_NAME=(config-map|secret):CONFIG_MAP_OR_SECRET_NAME ; you may provide this flag any number of times to define multiple volumes. To unset, specify the volume name followed by a "-" (e.g., /mount/path-).
      --volume-mount stringArray   Config a volume mount. /mount/path=VOLUME_MOUNT ; you may provide this flag any number of times to mount multiple volumes. To unset, specify the mount path followed by a "-" (e.g., /mount/path-).
      --wait-timeout int           Seconds to wait before giving up on waiting for service to be ready. (default 600)
```

### Options inherited from parent commands

```
      --config string       kn config file (default is $HOME/.kn/config.yaml)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --log-http            log http traffic
```

### SEE ALSO

* [kn service](kn_service.md)	 - Service command group

