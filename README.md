# Nova

Nova finds Helm charts in your Kubernetes cluster, and checks the upstream repository for any new releases.

# Installation

## From GitHub Releases
Visit the [releases page](https://github.com/FairwindsOps/nova/releases) to find the release
that's right for your environment. For example, on Linux:
```
curl -L "https://github.com/FairwindsOps/nova/releases/download/1.1.0/nova_1.1.0_linux_amd64.tar.gz" > nova.tar.gz
tar -xvf nova.tar.gz
sudo mv nova /usr/local/bin/
```

## Homebrew
```
brew tap fairwindsops/tap
brew install fairwindsops/tap/nova
```

## From source
```
go get https://github.com/fairwindsops/nova
```

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

