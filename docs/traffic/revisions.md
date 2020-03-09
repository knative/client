# Traffic Operations

`kn` supports the following traffic operations on traffic block of a Service as
part of `kn service update` command.

## Split Traffic

Every time you update the Service's template, a new Revision is created with
100% traffic routing to it. With every update to the Configuration of the
Service, a new Revision is created with Service route pointing all the traffic
to the latest ready Revision by default. However, we can change this behavior by
defining which Revision gets what portion of the traffic. For example, you might
want to carefully roll out 10% of traffic to your new Revision before putting
all traffic on it.

Updating the traffic is performed by `kn service update` with the `--traffic`
flag.

**`--traffic RevisionName=Percent`**

- `--traffic` flag requires two values separated by a `=`
- `RevisionName` string refers to name of Revision and `Percent` integer denotes
  the traffic portion to assign for this Revision
- `Percent` integer denotes the traffic portion to assign to the this Revision
- Use identifier `@latest` for RevisionName to refer to the latest ready
  Revision of the service, you can use this identifier only once with
  `--traffic` flag
- If the `service update` command is updating the Configuration values for the
  Service (along with traffic flags), `@latest` reference will be pointed to the
  created Revision after applying the update
- `--traffic` flag can be specified multiple times and is valid only if the sum
  of `Percent` values in all flags add up to 100
- Example: `kn service update svc --traffic @latest=10 --traffic svc-vwxyz=90`

### Tag Revisions

A tag in traffic block of Service creates a custom URL, which points to the
referenced Revision. A user can define a unique tag for an available Revision of
the Service thereby generating a custom URL of format
`http(s)://TAG-SERVICE.DOMAIN`. A given tag must be unique in traffic block of
the Service. `kn` supports assigning and unassigning custom tags for revisions
of the Service as part of `kn service update` command as follow:

**`--tag RevisionName=Tag`**

- `--tag` flag requires two values separated by a `=`
- `RevisionName` string refers to name of the `Revision`
- `Tag` string denotes the custom tag to be given for this Revision
- Use identifier `@latest` for RevisionName to refer to the latest ready
  Revision of the service, you can use this identifier only once with `--tag`
  flag
- If the `service update` command is updating the Configuration values for the
  Service (along with tag flags), `@latest` reference will be pointed to the
  created Revision after applying the update
- `--tag` flag can be specified multiple times
- `--tag` flag may assign different tags to same Revision, repeat the flag with
  same Revision name and different tags
- Example:
  `kn service update svc --tag @latest=candidate --tag svc-vwxyz=current`

_Note_:

- If you have assigned a tag to a particular Revision, a user can reference the
  Revision by its tag in `--traffic` flag as `--traffic Tag=Percent`.
- User can specify `--traffic` and `--tag` flags in same `service update`
  command and allowed to use the new tag for Revision reference for eg:
  `kn service update svc --tag echo-abcde=stable --traffic stable=100`.

### Untag Revisions:

Assigned tags to revisions in traffic block can be unassigned, thereby removing
any custom URLs. A user can unassign the tags for revisions as part of
`kn service update` command as follow:

**`--untag Tag`**

- `--untag` flag requires one value
- `tag` string denotes the unique tag in traffic block of the Service which
  needs to be unassigned and thereby also removing respective custom URL
- `--untag` flag can be specified multiple times
- Example: `kn service update svc --untag candidate`

**_NOTE_**:

- If a Revision is untagged and it is assigned 0% of the traffic, it is removed
  from the traffic block entirely.

## Flag operation precedence

- All the traffic-related flags can bespecified together in a single
  `kn service update` command, thus `kn` defines the precedence of these flags
  in which operations are evaluated.
- The order of the flags specified in the command are **not** taken into
  account.
- The precedence of flag(s) as they are evaluated by `kn`:

1. `--untag`: First, all the referenced tags by this flag are removed from the
   traffic block
2. `--tag`: Second, all the revisions are tagged as specified in the traffic
   block
3. `--traffic`: Third, all the referenced revisions (or tags) are assigned given
   portion of traffic split

**_NOTE_**:

- Use comma separated value pairs or separate value pairs: For example:
  `--tag echo-v1=v1,echo-v2=v2` is same as `--tag echo-v1=v1 --tag echo-v2=v2`
  similarly, `--traffic v1=50,v2=50` is same as
  `--traffic v1=50 --traffic v2=50` also, `--untag v1,v2` is same
  `--untag v1 --untag v2`
- `--untag` flag's action takes first precedence. If `--untag` and `--tag` flags
  specified together for revisions, all the mentioned tags will be unassigned
  from revisions first, followed by tagging them.
- Updating a tag for Revision requires untagging its existing tag and assigning
  a new tag. For example, to update tag from 'testing' to 'staging' for Revision
  'echo-v1': `--untag testing --tag staging`
- If a tag is assigned to a Revision and user tries to use same tag for another
  Revision, an error is raised.
- If a tag is assigned to a Revision and user tries to tag the same Revision
  with a new tag again, a new target is created with 0% traffic and new tag.
- To assign multiple tags to same Revision, do
  `--tag sameRevisionName=tagName1 --tag sameRevisionName=tagName2`.
- Repetition of `@latest` identifier for `--traffic` and `--tag` flags is not
  allowed.
- Repetition of revisionName or tagName for `--traffic` flag is not allowed;
  - If you want to have same revision present in multiple targets, use different
    tags to uniquely identify targets, and use unique tag names to assign
    traffic portions. For example:
    `kn service update svc --tag echo-v1=v1,echo-v2=v2, --traffic v1=50,v2=50`
- If one traffic block target has revisionName which matches tagName of another
  target, tagName is looked up first for assigning traffic. For example,
  `--traffic targetRef:percent` will assign mentioned `percent` to target where
  tag is `targetRef`. **Tags are unique** in traffic block, **RevisionNames are
  not**, use tag reference in such scenarios.

## Summary

Following table shows quick summary for traffic splitting flags, their value
formats, operation to perform and _Repetition_ column denotes if repeating the
particular value of flag is allowed in a `kn service update` command.

| Flag        |        Value(s)        | Operation                                           |        Repetition        |
| :---------- | :--------------------: | :-------------------------------------------------- | :----------------------: |
| `--traffic` | `RevisionName=Percent` | Give `Percent` traffic to `RevisionName`            |    :heavy_check_mark:    |
| `--traffic` |     `Tag=Percent`      | Give `Percent` traffic to the Revision having `Tag` |    :heavy_check_mark:    |
| `--traffic` |   `@latest=Percent`    | Give `Percent` traffic to the latest ready Revision | :heavy_multiplication_x: |
| `--tag`     |   `RevisionName=Tag`   | Give `Tag` to `RevisionName`                        |    :heavy_check_mark:    |
| `--tag`     |     `@latest=Tag`      | Give `Tag` to the latest ready Revision             | :heavy_multiplication_x: |
| `--untag`   |         `Tag`          | Remove `Tag` from Revision                          |    :heavy_check_mark:    |
