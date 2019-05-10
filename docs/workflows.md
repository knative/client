# Workflows

The purpose of this section of the Kn documentation is to list common workflows or use-cases for the Knative CLI. This is a live document, meant to be updated as we learn more about good ways to use `kn`.

## Basic

In this basic worflow we show the CRUD (create, read, update, delete) operations on a service. We use a well known [simple Hello World service](https://github.com/knative/docs/tree/master/docs/serving/samples/hello-world/helloworld-go) that reads the environment variable `TARGET` and prints it as output.

* Create a service from image in `default` namespace

```bash
$ kn service create hello --image gcr.io/knative-samples/helloworld-go --env TARGET=Knative
```

* Curl service endpoint

```bash
curl '-sS' '-H' 'Host: hello.default.example.com' 'http://xxx.xx.xxx.xx   '
Hello Knative!
```

Where `http://xxx.xx.xxx.xx` is your Knative installation ingress.

* List service

```bash
$ kn service list
$ hello
```

* Update service

```bash
$ kn service update hello --env TARGET=Kn
```

The service's environment variable `TARGET` is now set to `Kn`

<!-- TODO: include when `service get` is merged
* Get service

```bash
$ kn service get hello
#TODO ouput
```

TODO: observation(s) if any
-->

* Describe service

```bash
$ kn service describe hello
apiVersion: knative.dev/v1alpha1
kind: Service
metadata:
  creationTimestamp: "2019-05-09T21:14:41Z"
  generation: 1
  name: hello
  namespace: default
  resourceVersion: "28795267"
  selfLink: /apis/serving.knative.dev/v1alpha1/namespaces/default/services/hello
  uid: 760e1b9a-729f-11e9-9180-9ae5104abc98
spec:
  generation: 2
  runLatest:
    configuration:
      revisionTemplate:
        metadata:
          creationTimestamp: null
        spec:
          concurrencyModel: Multi
          container:
            env:
            - name: TARGET
              value: Kn
            name: ""
            resources: {}
status:
  conditions:
  - lastTransitionTime: "2019-05-09T21:14:52Z"
    status: "True"
    type: RoutesReady
  - lastTransitionTime: "2019-05-09T21:26:42Z"
    status: "True"
    type: ConfigurationsReady
  domain: hello.default.example.com
  domainInternal: hello.default.svc.cluster.local
  latestCreatedRevisionName: hello-00002
  latestReadyRevisionName: hello-00001
  observedGeneration: 2
  traffic:
  - configurationName: hello
    percent: 100
    revisionName: hello-00001
```

* Delete service

```bash
$ kn service delete hello
$ Deleted 'hello' in 'default' namespace.
```

You can then issue the same `$ kn service list` command to verify that the service was deleted.
