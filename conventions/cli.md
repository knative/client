# Guidelines for kn Commands

Commands are generally of the form `kn <resource> <verb>`.

## Resource

The <group>.knative.dev Kind, singluar and lowercase. For example, `service` for
`serving.knative.dev/service` or `trigger` for `eventing.knative.dev/trigger`.

## Verb

If the thing the user's doing can be described by the following commands, these
should be the name of the verb:

* `describe` prints detailed information about a single resource. It can include
  status information of related or child resources, too.

* `list` prints summary information about all resources of a type, possibly
  filtered by parent or label selector.

* `create` creates a resource. Accepts a `--force` flag to create-or-replace.

* `update` updates a resource based on the changes the user would like to make.

* `delete` deletes a resource

For a given resource there should be parallelism between arguments to `--create`
and `--update` as much as possible.

Other domain-specific verbs are possible, like `set-traffic` for a Knative
Serving Service.

## Arguments

### Positionals

Where there's a single target resource, it should be a positional argument.

```bash
kn service create foo --image gcr.io/things/stuff:tag
```
In this case `foo` is positional, and refers to the service to create.

### Flags

* `--force` is a flag on all create commands, and will replace the resource if
  it already exists (otherwise this is an error). The resource will be *mutated*
  to have a spec exactly like the resource that would otherwise be created. It
  is not deleted and recreated.

* When a flag sets a particular field on create or update, it should be a short
  name for the field, without necessarily specifying how it's nested. For
  example, `--image=img.repo/asdf` in Knative Serving sets
  `spec.template.containers[0].image`


