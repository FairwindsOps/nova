---
meta:
  - name: description
    content: "Fairwinds Nova | Documentation Quickstart"
---
# Quickstart

Install the golang binary and run it against your cluster.

```
$ go get github.com/fairwindsops/nova
$ nova find

Release Name      Installed    Latest     Old       Deprecated
============      =========    ======     ===       ==========
cert-manager      v0.11.0      v0.15.2    true      false
insights-agent    0.21.0       0.21.1     true      false
grafana           2.1.3        3.1.1      true      false
metrics-server    2.8.8        2.11.1     true      false
nginx-ingress     1.25.0       1.40.3     true      false
```

To check for outdated container images, instead of helm releases:

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
