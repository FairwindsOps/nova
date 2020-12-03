<div align="center">
  <img src="/img/logo.png" alt="Nova" />
  <br>

  <b>Find outdated or deprecated Helm charts running in your cluster.</b>

  [![GitHub release (latest SemVer)][release-image]][release-link] [![Version][version-image]][version-link] [![CircleCI][circleci-image]][circleci-link] [![Go Report Card][goreport-image]][goreport-link]
</div>

[version-image]: https://img.shields.io/github/go-mod/go-version/FairwindsOps/nova
[version-link]: https://github.com/FairwindsOps/nova

[release-image]: https://img.shields.io/github/v/release/FairwindsOps/nova
[release-link]: https://github.com/FairwindsOps/nova

[goreport-image]: https://goreportcard.com/badge/github.com/FairwindsOps/nova
[goreport-link]: https://goreportcard.com/report/github.com/FairwindsOps/nova

[circleci-image]: https://circleci.com/gh/FairwindsOps/nova.svg?style=svg
[circleci-link]: https://circleci.com/gh/FairwindsOps/nova


Nova scans your cluster for installed Helm charts, then cross-checks them against
all known Helm repositories. If it finds an updated version of the chart you're using,
or notices your current version is deprecated, it will let you know.

**Want to learn more?** Reach out on [Slack](https://fairwindscommunity.slack.com/) ([request invite](https://join.slack.com/t/fairwindscommunity/shared_invite/zt-e3c6vj4l-3lIH6dvKqzWII5fSSFDi1g)), send an email to `opensource@fairwinds.com`, or join us for [office hours on Zoom](https://fairwindscommunity.slack.com/messages/office-hours)

## Quickstart
```
$ go get github.com/fairwindsops/nova
$ nova find --helm-version=auto

Release Name      Installed    Latest     Old     Deprecated
cert-manager      v0.11.0      v0.15.2    true    false
insights-agent    0.21.0       0.21.1     true    false
grafana           2.1.3        3.1.1      true    false
metrics-server    2.8.8        2.11.1     true    false
nginx-ingress     1.25.0       1.40.3     true    false
```

## Installation

### From GitHub Releases
Visit the [releases page](https://github.com/FairwindsOps/nova/releases) to find the release
that's right for your environment. For example, on Linux:
```
curl -L "https://github.com/FairwindsOps/nova/releases/download/1.0.0/nova_1.0.0_linux_amd64.tar.gz" > nova.tar.gz
tar -xvf nova.tar.gz
sudo mv nova /usr/local/bin/
```

### Homebrew
```
brew tap fairwindsops/tap
brew install fairwindsops/tap/nova
```

### From source
```
go get github.com/fairwindsops/nova
```

## Usage
```
nova find --helm-version=auto --wide
```

### Options
* `--helm-version` - which version of Helm to use. Options are `2`, `3`, and `auto` (default is `3`)
* `--wide` - show `Chart Name`,  `Namespace` and `HelmVersion`
* `--output-file` - output JSON to a file
* `--url strings`, `-u` - URL for a helm chart repo (default [https://charts.fairwinds.com/stable,https://charts.fairwinds.com/incubator,https://kubernetes-charts.storage.googleapis.com,https://kubernetes-charts-incubator.storage.googleapis.com,https://charts.jetstack.io])
* `--poll-helm-hub` - When true, polls all helm repos that publish to helm hub (Default is true).
* `--helm-hub-config` - The URL to the helm hub sync config. (default is "https://raw.githubusercontent.com/helm/hub/master/config/repo-values.yaml")

### Output
Below is sample output for Nova

#### CLI
```bash
Release Name      Installed    Latest     Old     Deprecated
cert-manager      v0.11.0      v0.15.2    true    false
insights-agent    0.21.0       0.21.1     true    false
grafana           2.1.3        3.1.1      true    false
metrics-server    2.8.8        2.11.1     true    false
nginx-ingress     1.25.0       1.40.3     true    false
```

#### JSON
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
