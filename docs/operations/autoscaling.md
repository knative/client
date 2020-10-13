# Autoscaling

The Knative Pod Autoscaler (KPA), provides fast, request-based autoscaling
capabilities. To correctly configure autoscaling to zero for revisions, you must
modify its parameters.

`target` defines how many concurrent requests are wanted at a given time (soft
limit) and is the recommended configuration for autoscaling in Knative.

The `minScale` and `maxScale` annotations can be used to configure the minimum
and maximum number of pods that can serve applications.

You can access autoscaling capabilities by using `kn` to modify Knative services
without editing YAML files directly.

Use the `service create` and `service update` commands with the appropriate
flags to configure the autoscaling behavior.

| Flag                       | Description                                                                                                                 |
| :------------------------- | :-------------------------------------------------------------------------------------------------------------------------- |
| `--concurrency-limit int`  | Hard limit of concurrent requests to be processed by a single replica.                                                      |
| `--concurrency-target int` | Recommendation for when to scale up based on the concurrent number of incoming requests. Defaults to `--concurrency-limit`. |
| `--scale-max int`          | Maximum number of replicas.                                                                                                 |
| `--scale-min int`          | Minimum number of replicas.                                                                                                 |
