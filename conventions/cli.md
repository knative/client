## `kn` User Interface Principles

This document describes the conventions that are used for all `kn` commands and
options. It is normative in the sense that any new feature that introduces or
changes user-facing commands or options needs to adhere to the given principles.

Also, the given rules apply to plugins as well. Especially any plugin that wants
to be promoted to the
[knative/client-contrib](https://github.com/knative/client-contrib) plugin
repository has to adhere to these rules.

### Command Structure

The general format for kn and plugin commands is

```
kn <noun> [<noun2> ...] <verb> [<id> <id2] [--opt val1 --opt2 --opt3 ...]
```

So, commands are generally of the form `kn <noun> <verb>`
<sup>[1](#foot-1)</sup>, where [<noun>](#noun) is often the name of a resource
(e.g. `service`) but can also refer to other concepts (e.g. `plugin` or
`config`). This first noun forms a command group for all the operations you
might want to do with that kind of resource. Sometimes there can be deeper
hierarchies with multiple nouns (`kn <noun-1> <noun-2> .... <verb>`) when it
makes sense to structure complex concepts. A good example is
`kn source <source-type> <verb>` which is used like in `kn source ping create`.

`kn` commands take only positional arguments that are used as
[identifiers](#identifier). This identifier is often the name of a resource
which identifies the resource uniquely for the current or given namespace.

Top-level commands concerning the operation of `kn` itself, like `help` and
`version` are also okay.

### Noun

For resource-related commands, the kind itself used as a command in singular and
lowercase form. For example, `service` for `serving.knative.dev/service` or
`trigger` for `eventing.knative.dev/trigger` are the commands for managing these
resources respectively

### Verb

For CRUD (create-retrieve-update-delete) operation the following verbs have to
be used:

- `describe` prints detailed information about a single resource that can
  contain data of dependent objects, too.
- `list` prints summary information about all resources of a type.
- `create` creates a resource.
- `update` updates a resource.
- `delete` deletes a resource.
- `apply` for an idempotent "create-or-update", much like `kubectl apply`

For a given resource, create and update should use the same arguments as much as
possible and where it makes sense.

Other domain-specific verbs are possible on a case-by-case basis for operations
that go beyond basic CRUD operations.

### Identifier

For the `CRUD` operations `describe`, `create`, `update`, `delete` the
identifier is the resource's name and is required as a positional argument after
the commands. For example it is the last argument that does not start with a
flag prefix `-` or `--`. `list` operations can use a resource name to filter on
the resource.

Other identifiers can be plugin names or other entities' identifiers.

For bulk operations also multiple identifiers can be provided. For example, a
`delete` operation could use multiple resource names that should be deleted.

```bash
kn service create foo --image gcr.io/things/stuff:tag
```

In this case, `foo` is positional, and provides the name of the service to
create.

### Flags

Flags are used for specifying the input for `kn` commands and can have different
characteristics:

- They can be _mandatory_ or _optional_
- Mandatory flags are mentioned in the `Use` attribute of a command like in `service NAME --image IMAGE` for `ServiceCommand`
- Optional flags can have _default values_
- Flag values can be _scalars_, _binary_, _lists_ or _maps_ (see below for
  details)
- Flags always have a long-form (starting with a double `--`) but can also have
  a shortcut (beginning with a single `-`)
- Every flag has a help message attached
- Flags can be specific to a command or can be globally applicable

When adding new flags, the following recommendations should be considered:

- Never add a global flag except for very good reasons
- Group related flags together by using a common prefix, like `--label-revision`
  or `--label-service` so that they appear together in the help message (which
  is sorted alphabetically)
- Don't add a short form without former discussions
- Choose a name for the flag that is the same or close to the naming used in
  Knative serving itself like the corresponding CRD field or annotation name.

As mentioned above, flag values can be of different types. The rules of how
these values are modelled on the command line are given below.

#### Scalar

A scalar option is one which just takes a single value. This value can be a
string or a number. Such an option is allowed to be given only once. If given
multiple times, an error should be thrown.

A scalar flag's value can have an inner structure, too. For example
`--sink ksvc:myservice` uses a prefix `ksvc:` to indicate the targeted sink is a
Knative Service. A colon (`:`) should be used as separators if values have a
structure.

Example:

```
# Scalar parameter "--image" for specifying an application image
kn service create myservice --image docker.io/myuser/myimage
```

#### Binary

Binary flags come in pairs and don't carry any value. The flag representing the
`true` value is just the flag name without a value (e.g. `--wait`) whereas the
flag for a `false` value is this name with a `no-` prefix (e.g. `--no-wait`)

Example:

```
# Create a service an wait until deployed
kn service create myservice --wait ....

# Don't wait for the service to start
kn service create myservice --no-wait ...
```

Such a binary option can be provided only once. Otherwise, an error has to be
thrown.

#### List

List flag values can be provided in two flavours:

- Within a single flag value as comma-separated list of key-value pairs (e.g.
  `--resource pod:v1,job:batch/v1`)
- By providing the same option multiple times (e.g.
  `--resource pod:v1 --resource job:batch/v1`)

The value itself can carry a structure where colons separate the parts (`:`),
like in the examples above.

Example:

```
# Create an ApiServer source for listening on Pod and Job resource events
kn source apiserver create mysrc --resource pod:v1 --resource job:batch/v1 --sink ksvc:mysvc

# Same as above, but crammed into a single option
kn source apiserver create mysrc --resource pod:v1,job:batch/v1 --sink ksvc:mysvc
```

#### Maps

- Within a single flag value as comma separated list of key-value pairs (e.g.
  `--env USER=bla,PASSWORD=blub`)
- By providing the same option multiple times (e.g.
  `--env USER=bla --env PASSWORD=blub`)

For update operations, to _unset_ a value, the key has a dash suffix (`-`) and
no value part. For example, to _remove_ an environment variable named `USER`
from a service "hello-world".

If the same key is given multiple times on the command line, the latter
definition overwrites the previous one.

Example:

```
# Create a Service "hello-world" that sets USER and PASSWORD environment variables
kn service create hello-world --env USER=bla --env PASSWORD=blub

# Same as above
kn service create hello-world --env USER=bla,PASSWORD=blub

# Remove the USER envvar and add a HOME envvar to the service "hello-world"
kn service update hello-world --env USER- --env HOME=/root

# Same as above
kn service update hello-world --env USER-,HOME=/root

# Same as above, but the last HOME "/home" flag overwrites the previous one
kn service update hello-world --env HOME=/root --env USER-  --env HOME=/home
```

### Shared flags

Certain functionality is the same across command groups. For example, specifying
resource requests and limits via flags can be done when managing services but
also for sources. Those common functionalities should share the same
conventions, syntax.

Area to which this applies:

- Resource limits
- Output formats, i.e. the data formats supported by the `--output` option
  (which is reused from k8s' _cli-runtime_)
- Sinks
- ....

_this section needs to be completed with the concrete specifications. tbd_

<a name="foot-1">1</a>: Note that this differs from the `kubectl` model where
this order is vice versa (`kubectl <verb> <noun>`)
