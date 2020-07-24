<div align="center">
  <img src="/img/logo.png" alt="Nova" />
  <br>

  <b>Find outdated or deprecated Helm charts running in your cluster.</b>

  [![Version][version-image]][version-link] [![CircleCI][circleci-image]][circleci-link] [![Go Report Card][goreport-image]][goreport-link]
</div>

[version-image]: https://img.shields.io/static/v1.svg?label=Version&message=1.2.0&color=239922
[version-link]: https://github.com/FairwindsOps/nova

[goreport-image]: https://goreportcard.com/badge/github.com/FairwindsOps/nova
[goreport-link]: https://goreportcard.com/report/github.com/FairwindsOps/nova

[circleci-image]: https://circleci.com/gh/FairwindsOps/nova.svg?style=svg
[circleci-link]: https://circleci.com/gh/FairwindsOps/nova.svg


Nova scans your cluster for installed Helm charts, then cross-checks them against
all known Helm repositories. If it finds an updated version of the chart you're using,
or notices your current version is deprecated, it will let you know.

## Installation

### From GitHub Releases
Visit the [releases page](https://github.com/FairwindsOps/nova/releases) to find the release
that's right for your environment. For example, on Linux:
```
curl -L "https://github.com/FairwindsOps/nova/releases/download/1.1.0/nova_1.1.0_linux_amd64.tar.gz" > nova.tar.gz
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
go get https://github.com/fairwindsops/nova
```

## Usage

```
nova find --helm-version=auto
```

### Options
* `--helm-version` - which version of Helm to use. Options are `2`, `3`, and `auto` (default is `3`)
* `--wide` - show `Chart Name` and `Namespace`
* `--output-file` - output JSON to a file

### Output
Below is sample output for Nova

#### CLI
```
Release Name      Installed    Latest     Old     Deprecated
cert-manager      v0.11.0      v0.15.2    true    false
insights-agent    0.21.0       0.21.1     true    false
grafana           2.1.3        3.1.1      true    false
metrics-server    2.8.8        2.11.1     true    false
nginx-ingress     1.25.0       1.40.3     true    false
```

#### JSON
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

