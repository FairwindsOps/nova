---
meta:
  - name: description
    content: "Fairwinds Nova | Usage Documentation "
---
# Usage

```
nova find --wide
```

## Options
* `-h`, `--help` - help for nova
* `--config` - Pass a config file that can control the remaining settings. Command-line arguments still take precedence
* `--context` - Sets a specific context in the kubeconfig. If blank, uses the currently set context.
* `-d`, `--desired-versions` - A map of `chart=override_version` to override the helm repository when checking.
* `-a`, `--include-all` - Show all charts even if no latest version is found.
* `--output-file` - output JSON to a file
* `-v Level`, `--v Level` - set the log verbosity level where `Level` is a number between 1 and 10.
* `--wide` - show `Chart Name`,  `Namespace` and `HelmVersion`
* `--alsologtostderr` - log to standard error as well as files
* `--logtostderr` - log to standard error instead of files

## Generate Config

If you would like to generate a config file with all of the defaults for Nova, you can do that:

```
$ nova generate-config --config=nova.yaml
cat nova.yaml

context: ""
desired-versions: {}
include-all: false
output-file: ""
wide: false
```

## Output
Below is sample output for Nova

### CLI
```bash
Release Name      Installed    Latest     Old     Deprecated
============      =========    ======     ===     ==========
goldilocks        3.3.1        4.0.1      true    false
metrics-server    5.6.0        5.10.10    true    false
redis             15.4.1       15.5.5     true    false
```

### CLI (with --wide)

```
Release Name      Chart Name        Namespace         HelmVersion    Installed    Latest     Old     Deprecated
============      ==========        =========         ===========    =========    ======     ===     ==========
goldilocks        goldilocks        goldilocks        3              3.3.1        4.0.1      true    false
metrics-server    metrics-server    metrics-server    3              5.6.0        5.10.10    true    false
redis             redis             redis             3              15.4.1       15.5.5     true    false
```

### JSON
```json
{
    "helm": [
        {
          "release": "goldilocks",
          "chartName": "goldilocks",
          "namespace": "goldilocks",
          "description": "A Helm chart for running Fairwinds Goldilocks. See https://github.com/FairwindsOps/goldilocks\n",
          "icon": "https://raw.githubusercontent.com/FairwindsOps/charts/master/stable/goldilocks/icon.png",
          "Installed": {
            "version": "3.3.1",
            "appVersion": "v3.1.4"
          },
          "Latest": {
            "version": "4.0.1",
            "appVersion": "v4.0.0"
          },
          "outdated": true,
          "deprecated": false,
          "helmVersion": "3",
          "overridden": false
        }
    ]
}
```
