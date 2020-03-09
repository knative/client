# Guidelines for `kn` Commands

Commands are generally of the form `kn <resource> <verb>`; the resource kind
forms a command group for all the operations you might want to do with that kind
of resource.

Commands that directly concern more than one resource kind may be categorized
with one of the relevant resources, or may get their own top-level verb (eg.
`connect`).

Top-level commands concerning the operation of `kn` itself, like `help` and
`version` are also okay.

## Resource

The <group>.knative.dev Kind, singluar and lowercase. For example, `service` for
`serving.knative.dev/service` or `trigger` for `eventing.knative.dev/trigger`.

## Verb

If the thing the user's doing can be described by the following commands, these
should be the name of the verb:

- `describe` prints detailed information about a single resource. It can include
  status information of related or child resources, too.

- `list` prints summary information about all resources of a type, possibly
  filtered by parent or label selector.

- `create` creates a resource. Accepts a `--force` flag to create-or-replace.

- `update` updates a resource based on the changes the user would like to make.

- `delete` deletes a resource

For a given resource there should be parallelism between arguments to `create`
and `update` as much as possible.

Other domain-specific verbs are possible on a case-by-case basis, like
`set-traffic` for a Knative Serving Service.

## Arguments

### Positionals

Where there's a single target resource, the resource name should be a positional
argument. It needs to be of the resource type we're talking about, eg.
`kn revision` subcommands the positional must be naming a revision.

```bash
kn service create foo --image gcr.io/things/stuff:tag
```

In this case `foo` is positional, and refers to the service to create.

### Flags

- `--force` is a flag on all create commands, and will replace the resource if
  it already exists (otherwise this is an error). The resource will be _mutated_
  to have a spec exactly like the resource that would otherwise be created. It
  is not deleted and recreated.

- When a flag sets a particular field on create or update, it should be a short
  name for the field, without necessarily specifying how it's nested. For
  example, `--image=img.repo/asdf` in Knative Serving sets
  `spec.template.containers[0].image`

- Flags that control a boolean behavior (eg. generate a name or not) are
  specified by their presence. The default happens when the flag is not present,
  and adding the flag marks the user's desire for the non-default thing. When
  the flag _disables_ a default behavior which is to do something, it should
  start with `no` (eg. `--no-generate-name` when the default is to generate a
  name).

#### Output

Commands that output information should support `--output` with a shorthand of
`-o` to choose how to frame their output, and `--template` for supplying
templates to output styles that use them.
