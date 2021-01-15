# Setting Desired Versions

If you would like to specify your own set of versions, rather than using the latest available version for comparison, you may set `desired-versions` via the command-line or via file.

## Using a Config File

nova.yaml
```yaml
desired-versions:
  metrics-server: 6.0.0
```

Example run with config:
```
$ nova find --config nova.yaml  --poll-helm-hub=false
Release Name      Installed    Latest    Old     Deprecated
metrics-server    5.3.3        6.0.0     true    false
```

Then again, without the config and polling helm hub:
```
$ nova find
Release Name      Installed    Latest    Old      Deprecated
metrics-server    5.3.3        5.3.3     false    false
```

## Using the CLI

```
$ nova find --desired-versions='metrics-server=6.0.0,vpa=12.0.0'
Release Name      Installed    Latest    Old     Deprecated
metrics-server    5.3.3        6.0.0     true    false
vpa               0.2.2        12.0.0    true    false
```
