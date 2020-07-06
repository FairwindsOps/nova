# Nova

Nova finds Helm charts in your Kubernetes cluster, and checks the upstream repository for any new releases.

# Usage

```
nova find --output-file=releases.json
```

# Output
Below is sample output for Nova

## CLI
```
ReleaseName               ChartName                 Namespace            Version       NewestVersion  IsOld    Deprecated
cert-manager              cert-manager              cert-manager         v0.11.0       v0.15.2        True
insights-agent            insights-agent            insights-agent       0.21.0        0.21.0
grafana                   grafana                   insights-tools-graf… 2.1.3         3.0.1          True
metrics-server            metrics-server            metrics-server       2.8.8         2.11.1         True
nginx-ingress             nginx-ingress             nginx-ingress        1.25.0        1.40.2         True
```

## JSON
```
{
    "helm_releases": [
        {
            "release": "cert-manager",
            "chartName": "cert-manager",
            "namespace": "cert-manager",
            "description": "A Helm chart for cert-manager",
            "home": "https://github.com/jetstack/cert-manager",
            "icon": "https://raw.githubusercontent.com/jetstack/cert-manager/master/logo/logo.png",
            "version": "v0.11.0",
            "appVersion": "v0.11.0",
            "newest": "v0.15.2",
            "newest_appVersion": "v0.15.2",
            "outdated": true
        }
    ]
}
```

