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
```
Flags:
      --containers        Show old container image versions. There will be no helm output unless the --helm flag is set as well
      --helm              Show old helm releases. You can combine this flag with `--containers` to have both output in a single run.
  -h, --help              help for find
      --show-errored-containers   When finding container images, show errors encountered when scanning.
      --show-non-semver           When finding container images, show all containers even if they don't follow semver.
  -t, --timeout uint16            When finding container images, the time in seconds before canceling the operation. (default 10)

Global Flags:
      --alsologtostderr                   log to standard error as well as files (default true)
      --config string                     Config file to use. If empty, flags will be used instead
      --context string                    A context to use in the kubeconfig.
  -d, --desired-versions stringToString   A map of chart=override_version to override the helm repository when checking. (default [])
  -a, --include-all                       Show all charts even if no latest version is found.
      --logtostderr                       log to standard error instead of files (default true)
  -n, --namespace string                  Namespace to look in. If empty, scan will be cluster-wide
      --output-file string                Path on local filesystem to write file output to
      --poll-artifacthub                  When true, polls artifacthub to match against helm releases in the cluster. If false, you must provide a url list via --url/-u. Default is true. (default true)
      --show-old                          Only output charts that are not on the latest version
  -u, --url strings                       URL for a helm chart repo
  -v, --v Level                           number for the log level verbosity
      --wide                              Output chart name and namespace
```

## Referencing Private Registries

If you would like to use nova to find outdated releases from charts in private helm repositories, please use the `--url` flag to point to that registry. Any authentication should be already configured in your local `helm repo` settings. Additionally, you may want to set `--poll-artifacthub=false` if there are no releases from public repositories that you wish to find.

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

## Helm Scanning Output
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
Release Name      Chart Name        Namespace         HelmVersion   KubeVersion    Installed    Latest   Old  Deprecated
============      ==========        =========         ===========   ===========    ===========  ======   ===  ==========
cert-manager     cert-manager      cert-manager          3         >= 1.22.0-0      v.1.12.0   v.1.12.2 true    false
vault             vault             default               3         >= 1.20.0-0      0.25.0     0.25.0   false   false
goldilocks        goldilocks        goldilocks            3                          7.0.0      7.0.0    false   false
```

### JSON
```json
  {
    "release": "cert-manager",
    "chartName": "cert-manager",
    "namespace": "cert-manager",
    "description": "A Helm chart for cert-manager",
    "home": "https://github.com/cert-manager/cert-manager",
    "icon": "https://raw.githubusercontent.com/cert-manager/cert-manager/d53c0b9270f8cd90d908460d69502694e1838f5f/logo/logo-small.png",
    "Installed": {
      "version": "v1.12.0",
      "appVersion": "v1.12.0"
    },
    "Latest": {
      "version": "1.12.2",
      "appVersion": "v1.12.2"
    },
    "outdated": true,
    "deprecated": false,
    "helmVersion": "3",
    "kubeVersion": ">= 1.22.0-0",
    "overridden": false
  }
```

## Container Image Output
There are a couple flags that are unique to the container image output.
- `--show-non-semver` will also show any container tags running in the cluster that do not have valid semver versions. By default these are not shown.
- `--show-errored-containers` will show any containers that returned some sort of error when reaching out to the registry and/or when processing the tags.
- `--timeout` will set the time (in seconds) before remote queries to the registry are cancelled. Useful when an image has many tags. Defaults to 10 seconds.

Below is sample output for Nova when using the `--containers` flag

```
$ nova find --containers
Container Name                              Current Version    Old     Latest     Latest Minor     Latest Patch
==============                              ===============    ===     ======     =============    =============
k8s.gcr.io/coredns/coredns                  v1.8.0             true    v1.8.6     v1.8.6           v1.8.6
k8s.gcr.io/etcd                             3.4.13-0           true    3.5.3-0    3.4.13-0         3.4.13-0
k8s.gcr.io/kube-apiserver                   v1.21.1            true    v1.23.6    v1.23.6          v1.21.12
k8s.gcr.io/kube-controller-manager          v1.21.1            true    v1.23.6    v1.23.6          v1.21.12
k8s.gcr.io/kube-proxy                       v1.21.1            true    v1.23.6    v1.23.6          v1.21.12
k8s.gcr.io/kube-scheduler                   v1.21.1            true    v1.23.6    v1.23.6          v1.21.12
```

### Container output with errors
When scanning all containers, nova will capture any errors and move on. To show which containers had errors, use the `--show-errored-containers` flag. Output will look like:

```
$ nova find --containers --show-errored-containers
Container Name                              Current Version    Old     Latest     Latest Minor     Latest Patch
==============                              ===============    ===     ======     =============    =============
k8s.gcr.io/coredns/coredns                  v1.8.0             true    v1.8.6     v1.8.6           v1.8.6
k8s.gcr.io/etcd                             3.4.13-0           true    3.5.3-0    3.4.13-0         3.4.13-0
k8s.gcr.io/kube-apiserver                   v1.21.1            true    v1.23.6    v1.23.6          v1.21.12
k8s.gcr.io/kube-controller-manager          v1.21.1            true    v1.23.6    v1.23.6          v1.21.12
k8s.gcr.io/kube-proxy                       v1.21.1            true    v1.23.6    v1.23.6          v1.21.12
k8s.gcr.io/kube-scheduler                   v1.21.1            true    v1.23.6    v1.23.6          v1.21.12


Errors:
Container Name                        Error
==============                        =====
examplething.com/testing:v1.0.0       Get "https://examplething.com/v2/": dial tcp: lookup examplethingert.com: no such host                                                                                                 =====
```

## Helm Releases and Container Images combined output
If you want to run nova and both helm releases and containers images results in a single run

Below is sample output for Nova when using the `--helm --containers` flag

```
$ nova --format=table find --helm --containers
Release Name      Installed    Latest    Old      Deprecated
============      =========    ======    ===      ==========
cert-manager      v1.9.1       1.9.1     false    false
insights-agent    2.0.7        2.6.8     true     false

Container Name                        Current Version    Old     Latest     Latest Minor     Latest Patch
==============                        ===============    ===     ======     =============    =============
k8s.gcr.io/coredns/coredns            v1.8.4             true    v1.9.3     v1.9.3           v1.8.6
k8s.gcr.io/etcd                       3.5.0-0            true    3.5.4-0    3.5.0-0          3.5.0-0
k8s.gcr.io/kube-apiserver             v1.22.9            true    v1.25.0    v1.25.0          v1.22.13
k8s.gcr.io/kube-controller-manager    v1.22.9            true    v1.25.0    v1.25.0          v1.22.13
k8s.gcr.io/kube-proxy                 v1.22.9            true    v1.25.0    v1.25.0          v1.22.13
k8s.gcr.io/kube-scheduler             v1.22.9            true    v1.25.0    v1.25.0          v1.22.13
```

You can print the output in `json` format

```
$ nova --format=json find --helm --containers | jq
{
  "helm": [
    {
      "release": "cert-manager",
      "chartName": "cert-manager",
      "namespace": "cert-manager",
      "description": "A Helm chart for cert-manager",
      "home": "https://github.com/cert-manager/cert-manager",
      "icon": "https://raw.githubusercontent.com/cert-manager/cert-manager/d53c0b9270f8cd90d908460d69502694e1838f5f/logo/logo-small.png",
      "Installed": { "version": "v1.9.1", "appVersion": "v1.9.1" },
      "Latest": { "version": "1.9.1", "appVersion": "v1.9.1" },
      "outdated": false,
      "deprecated": false,
      "helmVersion": "3",
      "kubeVersion": "",
      "overridden": false
    }
  ],
  "include_all": false,
  "container": {
    "container_images": [
      {
        "name": "k8s.gcr.io/kube-scheduler",
        "current_version": "v1.22.9",
        "latest_version": "v1.25.0",
        "latest_minor_version": "v1.25.0",
        "latest_patch_version": "v1.22.13",
        "outdated": true,
        "affectedWorkloads": [
          {
            "name": "kube-scheduler-kind-control-plane",
            "namespace": "kube-system",
            "kind": "Pod",
            "container": "kube-scheduler"
          }
        ]
      }
    ],
    "err_images": null,
    "latest_string_found": false
  }
}
```
