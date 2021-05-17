# Traffic Splitting Examples

1. Tag revisions `echo-v1` and `echo-v2` as `stable` and `staging` respectively:

```
kn service update svc --tag echo-v1=stable --tag echo-v2=staging
```

2. Ramp up/down revision `echo-v3` to `20%`, adjusting other traffic to
   accommodate:

```
kn service update svc --traffic echo-v3=20 --traffic echo-v2=80
```

3. Give revision `echo-v3` tag `candidate`, without otherwise changing any
   traffic split:

```
kn service update svc --tag echo-v3=candidate
```

4. Give `echo-v3` tag `candidate`, and `2%` of traffic adjusting other traffic
   to go to revision `echo-v2`:

```
kn service update svc --tag echo-v3=candidate --traffic candidate=2 --traffic echo-v2=98
```

6. Update tag for `echo-v3` from `candidate` to `current`:

```
kn service update svc --untag candidate --tag echo-v3=current
```

7. Remove tag `current` from `echo-v3`:

```
kn service update svc --untag current
```

8. Remove `echo-v3` having no tag(s) entirely, adjusting `echo-v2` to fill up:

```
kn service update svc --traffic echo-v2=100    # a target having no-tags or no-traffic gets removed
```

9. Remove `echo-v1` and its tag `old` from the traffic assignments entirely:

```
kn service update svc --untag old --traffic echo-v1=0
```

10. Tag revision `echo-v1` with `stable` as well as `current`, and `50-50%`
    traffic split to each:

```
kn service update svc --tag echo-v1=stable,echo-v2=current --traffic stable=50,current=50
```

11. Revert all the traffic to the latest ready revision of service:

```
kn service update svc --traffic @latest=100
```

12. Tag latest ready revision of service as `current`:

```
kn service update svc --tag @latest=current
```

13. Update tag for `echo-v4` to `testing` and assign all traffic to it:

```
kn service update svc --untag oldv4 --tag echo-v4=testing --traffic testing=100
```

14. Update `latest` tag of `echo-v1` with `old` tag, give `latest` to `echo-v2`:

```
kn service update svc --untag latest --tag echo-v1=old --tag echo-v2=latest
```
