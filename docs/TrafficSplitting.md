# Traffic Splitting

The purpose of this section of the kn documentation is to describe the command verbs for traffic splitting operations and present some examples.

`kn` client can work with following properties of a traffic target defined in traffic block of a service spec.
 - **RevisionName**: RevisionName of a specific revision to which to send a portion of traffic.
 - **Tag**: Tag is optionally used to expose a dedicated url for referencing this target exclusively.
 - **Percent**: Percent specifies percent of the traffic to this Revision.

## Operations

Operations required to perform traffic splitting use cases, for example: blue/green deployment, etc can be summarized as below:
 - Tag targets
 - Split traffic

### Tag Targets

User can tag traffic targets (revisions) to
 - Create custom URLs
 - Create a human-friendly reference to refer to a traffic target
 - Create static routes and change the underlying revisions as needed
 - Remove the custom URLs OR remove the tags of targets
 - Remove a target from traffic block altogether


#### Tagging commands

#### 1. `kn service tag-target [RevisionName1:Tag1] [RevisionName2:Tag2]`

 - Tags given revision with specified tag thereby creating a custom URL
 - Does not change the traffic percent of the target
 - Multiple `RevisionName:Tag` combination can be specified
 - Multiple tags for same revision can be specified
   - `kn service tag-target echo-v1:stable echo-v1:current`
 - If the target doesn't exist in deployed service, it creates a target entry with 0% traffic portion
 - For updating tags of existing traffic targets, use `kn service untag-target` followed by `kn service tag-target`, because
    - The deployed service might have multiple tags referring to a `RevisionName`
    - The specified `RevisionName:Tag` combination on CLI might request to have different tags for same revision

#### 2. `kn service untag-target [RevisionName:Tag] [RevisionName [--all]] [Tag]`

 - Takes a single reference of a traffic target
 - Does not change the traffic percent of the target
 - Can be specified either via
   -  `RevisionName`, removes the tag for target where `RevisionName` matches, errors if multiple tags for a revision exists
   -  `Tag`, removes the specified tag from traffic targets
   -  `RevisionName:Tag` removes `Tag` for `RevisionName`, useful for case where a single revision is tagged with multiple tags
   - `--all` flag can be optionally specified along with `RevisionName` to remove all tags for given revision
 - If a target has 0% traffic portion and it is untagged, target is removed from traffic block altogether


## Split Traffic

A user can
 - Specify the traffic percent portion to route to a particular target
 - Adjust the traffic percent of existing traffic targets
 - Reference a traffic target either by `RevisionName` or `Tag`

### Traffic splitting command

#### `kn service set-traffic serviceName [revisionName1=Percent] [Tag2=Percent] [Tag3=*]`

 - Takes the first argument as service name to operate on
 - Subsequent arguments specifies the traffic targets
 - Traffic target can be referenced either via `RevisionName` or `Tag`
 - Traffic portion to allocate is specified after `=`
 - Wildcard character `*` can be specified to allocate portion remaining after, summing specified portions and substracting it from 100
 - Specified traffic allocations replace existing traffic block of deployed service

#### Referencing `latestRevision:true` (`$`) target
 - Serving allows having `latestRevision:true` field in traffic block as traffic target
 - This field indicates that the latest ready Revision of the Configuration should be used for this traffic target
 - We can refer this as placeholder to point to the latest Revision which will be generated for the Service
 - To reference this special target in traffic block, we can use character `$`, which can not be a Revision name to avoid conflicts
 - Referencing `$` to tag a target or set traffic percent will update target in traffic block with Revision reference to `latestRevision: true`

```
kn service tag-target $:current
```
will form the traffic block as below before updating service
```
  targets:
  - latestRevision: true
    tag: current
    percent: xx
```

and,
```
kn service set-traffic svc1 $:100   #OR 'kn service set-traffic svc1 current:100'
```
will form the traffic block
```
  targets:
  - latestRevision: true
    tag: current
    percent: 100
```


## Examples:

1. Tag revisions `echo-v1` and `echo-v2` as `stable` and `staging`
```
kn service tag-target echo-v1:stable echo-v2:staging
```

2. Ramp up/down revision echo-v3 to 20%, adjusting other traffic to accommodate:
```
kn service set-traffic svc1 echo-v3=20 echo-v2=*
```

3. Give revision echo-v3 the tag candidate, without otherwise changing any traffic split:
```
kn service tag-target echo-v3:candidate
```

4. Give echo-v3 the tag candidate, and 2% of traffic adjusting other traffic to go to revision echo-v2:
```
kn service tag-target echo-v3:candidate
kn service set-traffic svc1 candidate=2 echo-v2=*
```

5. Give whatever revision has the tag candidate 10% of the traffic, adjusting the traffic on whatever revision has the tag current to accommodate:
```
kn service set-traffic svc1 candidate=10 current=*
```

6. Update the tag for echo-v3 from candidate to current:
```
kn service untag-target echo-v3:candidate
kn service tag-target echo-v3 current
```

7. Remove the tag current from echo-v3:
```
kn service untag-target echo-v3:current
```

8. Remove echo-v3 from the traffic assignments entirely, adjusting echo-v2 to fill up:
```
kn service set-traffic svc1 echo-v3=0 echo-v2=*  #echo-v2 gets 100% traffic portion here
kn service untag-target echo-v3:current
```

9. Tag revision echo-v1 as stable and current with 50-50% traffic split
```
kn service tag-target echo-v1:stable echo-v1:current
kn service set-traffic svc1 stable=50 current=50
```

10. Revert all the traffic to latest revision of service
```
kn service set-traffic svc1 $=100
```
