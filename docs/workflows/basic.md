# Basic Workflow

In this basic workflow we show the CRUD (create, read, update, delete) operations
on a service. We use a well known
[simple Hello World service](https://github.com/knative/docs/tree/master/docs/serving/samples/hello-world/helloworld-go)
that reads the environment variable `TARGET` and prints it as output.

- **Create a service in the `default` namespace from an image**

```bash
kn service create hello --image gcr.io/knative-samples/helloworld-go --env TARGET=Knative

Creating service 'hello' in namespace 'default':

  0.247s The Route is still working to reflect the latest desired specification.
  0.299s Configuration "hello" is waiting for a Revision to become ready.
 11.631s ...
 11.719s Ingress has not yet been reconciled.
 13.102s Ready to serve.

Service 'hello' created with latest revision 'hello-bxshg-1' and URL:
http://hello.default.apps-crc.testing
```

- **List a service**

```bash
kn service list
NAME    URL                                LATEST          AGE     CONDITIONS   READY   REASON
hello   http://hello.default.example.com   hello-dskww-1   2m42s   3 OK / 3     True
```

- **Curl service endpoint**

```bash
curl '-sS' '-H' 'Host: hello.default.example.com' 'http://xxx.xx.xxx.xx   '
Hello Knative!
```

Where `http://xxx.xx.xxx.xx` is your Knative installation ingress.

- **Update a service**

```bash
kn service update hello --env TARGET=Kn

Updating Service 'hello' in namespace 'default':

  3.559s Traffic is not yet migrated to the latest revision.
  3.624s Ingress has not yet been reconciled.
  3.770s Ready to serve.

Service 'hello' updated with latest revision 'hello-nhbwv-2' and URL:
http://hello.default.example.com
```

The service's environment variable `TARGET` is now set to `Kn`.

- **Describe a service**

```bash
kn service describe hello
Name:       hello
Namespace:  default
Age:        5m
URL:        http://hello.default.example.com
Address:    http://hello.default.svc.cluster.local

Revisions:
  100%  @latest (hello-nhbwv-2) [2] (50s)
        Image:  gcr.io/knative-samples/helloworld-go (pinned to 5ea96b)

Conditions:
  OK TYPE                   AGE REASON
  ++ Ready                  46s
  ++ ConfigurationsReady    46s
  ++ RoutesReady            46s
```

- **Delete a service**

```bash
kn service delete hello
Service 'hello' successfully deleted in namespace 'default'.
```

You can then verify that the 'hello' service is deleted by trying to `list` it
again.

```bash
kn service list hello
No services found.
```
