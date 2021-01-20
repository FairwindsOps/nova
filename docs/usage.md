# Usage

```
nova find --helm-version=auto --wide
```

## Options
* `--config` - Pass a config file that can control the remaining settings. Command-line arguments still take precedence
* `--helm-version` - which version of Helm to use. Options are `2`, `3`, and `auto` (default is `3`)
* `--context` - Sets a specific context in the kubeconfig. If blank, uses the currently set context.
* `-d`, `--desired-versions` - A map of `chart=override_version` to override the helm repository when checking.
* `--wide` - show `Chart Name`,  `Namespace` and `HelmVersion`
* `--output-file` - output JSON to a file
* `--url strings`, `-u` - URL for a helm chart repo (default [https://charts.fairwinds.com/stable,https://charts.fairwinds.com/incubator,https://kubernetes-charts.storage.googleapis.com,https://kubernetes-charts-incubator.storage.googleapis.com,https://charts.jetstack.io])
* `--poll-helm-hub` - When true, polls all helm repos that publish to helm hub (Default is true).
* `--helm-hub-config` - The URL to the helm hub sync config. (default is "https://raw.githubusercontent.com/helm/hub/master/config/repo-values.yaml")

## Generate Config

If you would like to generate a config file with all of the defaults for Nova, you can do that:

```
$ nova generate-config --config=nova.yaml
cat nova.yaml

context: ""
desired-versions: {}
helm-hub-config: https://raw.githubusercontent.com/helm/hub/master/config/repo-values.yaml
helm-version: "3"
output-file: ""
poll-helm-hub: true
url:
- https://charts.fairwinds.com/stable
- https://charts.fairwinds.com/incubator
- https://kubernetes-charts.storage.googleapis.com
- https://kubernetes-charts-incubator.storage.googleapis.com
- https://charts.jetstack.io
wide: false
```

## Output
Below is sample output for Nova

### CLI
```bash
Release Name      Installed    Latest     Old     Deprecated
cert-manager      v0.11.0      v0.15.2    true    false
insights-agent    0.21.0       0.21.1     true    false
grafana           2.1.3        3.1.1      true    false
metrics-server    2.8.8        2.11.1     true    false
nginx-ingress     1.25.0       1.40.3     true    false
```

### CLI (with --wide)

```
Release Name      Chart Name        Namespace         HelmVersion    Installed    Latest    Old      Deprecated
metrics-server    metrics-server    metrics-server    v3             5.3.3        5.3.3     false    false
vpa               vpa               vpa               v3             0.2.2        0.2.2     false    false
```

### JSON
```json
{
    "helm": [
        {
            "release": "cert-manager",
            "chartName": "cert-manager",
            "namespace": "cert-manager",
            "description": "A Helm chart for cert-manager",
            "home": "https://github.com/jetstack/cert-manager",
            "icon": "https://raw.githubusercontent.com/jetstack/cert-manager/master/logo/logo.png",
            "Installed": {
                "version": "v0.11.0",
                "appVersion": "v0.11.0"
            },
            "Latest": {
                "version": "v0.16.0",
                "appVersion": "v0.16.0"
            },
            "helmVersion": "v3",
            "outdated": true,
            "deprecated": false
        }
    ]
}
```
