# Workflows

The purpose of this section of the Kn documentation is to list common workflows or use-cases for the Knative CLI. This is a live document, meant to be updated as we learn more about good ways to use `kn`.

## Basic

In this basic worflow we show the CRUD (create, read, update, delete) operations on a service. We use a well known [simple Hello World service](https://github.com/knative/docs/tree/master/docs/serving/samples/hello-world/helloworld-go) that reads the environment variable `TARGET` and prints it as output.

* **Create a service from image in `default` namespace**

```bash
kn service create hello --image gcr.io/knative-samples/helloworld-go --env TARGET=Knative
Service 'hello' successfully created in namespace 'default'.
```

* **Get service**

```bash
kn service get hello
NAME           DOMAIN                             GENERATION   AGE     CONDITIONS   READY   REASON
hello          hello.default.example.com          1            3m5s    3 OK / 3     True
```

* **Curl service endpoint**

```bash
curl '-sS' '-H' 'Host: hello.default.example.com' 'http://xxx.xx.xxx.xx   '
Hello Knative!
```

Where `http://xxx.xx.xxx.xx` is your Knative installation ingress.

* **Update service**

```bash
kn service update hello --env TARGET=Kn
```

The service's environment variable `TARGET` is now set to `Kn`.

* **Describe service**

```bash
kn service describe hello
```
```yaml
apiVersion: knative.dev/v1alpha1
kind: Service
metadata:
  creationTimestamp: "2019-05-14T20:11:06Z"
  generation: 1
  name: hello
  namespace: default
  resourceVersion: "29659961"
  selfLink: /apis/serving.knative.dev/v1alpha1/namespaces/default/services/hello
  uid: 67d46126-7684-11e9-b088-4639f5970760
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
  - lastTransitionTime: "2019-05-14T20:11:42Z"
    status: "True"
    type: RoutesReady
  - lastTransitionTime: "2019-05-14T20:14:53Z"
    status: Unknown
    type: ConfigurationsReady
  - lastTransitionTime: "2019-05-14T20:14:53Z"
    status: Unknown
    type: Ready
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

* **Delete service**

```bash
kn service delete hello
Service 'hello' successfully deleted in namespace 'default'.
```

You can then verify that the 'hello' service is deleted by trying to `get` it again.

```bash
kn service get hello
No resources found.
```