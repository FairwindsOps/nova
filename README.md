<div align="center" class="no-border">
  <img src="/img/logo.png" alt="Nova" />
  <br>

  <b>Find outdated or deprecated Helm charts running in your cluster.</b>

  <a href="https://github.com/FairwindsOps/nova/releases">
    <img src="https://img.shields.io/github/v/release/FairwindsOps/nova">
  </a>
  <a href="https://goreportcard.com/report/github.com/FairwindsOps/nova">
    <img src="https://goreportcard.com/badge/github.com/FairwindsOps/nova">
  </a>
  <a href="https://circleci.com/gh/FairwindsOps/nova.svg">
    <img src="https://circleci.com/gh/FairwindsOps/nova.svg?style=svg">
  </a>
</div>

Nova scans your cluster for installed Helm charts, then cross-checks them against
all known Helm repositories. If it finds an updated version of the chart you're using,
or notices your current version is deprecated, it will let you know.

Nova can also scan your cluster for out of date container images. Find out more in the [docs](https://nova.docs.fairwinds.com).

## Documentation

Check out the [documentation at docs.fairwinds.com](https://nova.docs.fairwinds.com)

## Notice: Registry Migration and Immutable Images (v3.11.15 → v3.12.0)

Starting with **v3.12.0**:

- Images moved to `us-docker.pkg.dev/fairwinds-ops/oss/nova`
- `quay.io/fairwinds/nova` is deprecated

### Required action

```diff
- quay.io/fairwinds/nova:<tag>
+ us-docker.pkg.dev/fairwinds-ops/oss/nova:<tag>
```

---

## Immutable and signed images

* Images are now **signed**
* Tags are **immutable**
* No more floating tags:

  * `v3`
  * `v3.11`
  * `latest`

Use full version tags:

```
us-docker.pkg.dev/fairwinds-ops/oss/nova:v<major>.<minor>.<patch>
```

Or pin by digest:

```
us-docker.pkg.dev/fairwinds-ops/oss/nova@sha256:<digest>
```


<!-- Begin boilerplate -->
## Join the Fairwinds Open Source Community

The goal of the Fairwinds Community is to exchange ideas, influence the open source roadmap,
and network with fellow Kubernetes users.
[Chat with us on Slack](https://join.slack.com/t/fairwindscommunity/shared_invite/zt-2na8gtwb4-DGQ4qgmQbczQyB2NlFlYQQ)

## Other Projects from Fairwinds

Enjoying Nova? Check out some of our other projects:
* [Polaris](https://github.com/FairwindsOps/Polaris) - Audit, enforce, and build policies for Kubernetes resources, including over 20 built-in checks for best practices
* [Goldilocks](https://github.com/FairwindsOps/Goldilocks) - Right-size your Kubernetes Deployments by compare your memory and CPU settings against actual usage
* [Pluto](https://github.com/FairwindsOps/Pluto) - Detect Kubernetes resources that have been deprecated or removed in future versions
* [rbac-manager](https://github.com/FairwindsOps/rbac-manager) - Simplify the management of RBAC in your Kubernetes clusters

Or [check out the full list](https://www.fairwinds.com/open-source-software?utm_source=nova&utm_medium=nova&utm_campaign=nova)
## Fairwinds Insights
If you're interested in running Nova in multiple clusters,
tracking the results over time, integrating with Slack, Datadog, and Jira,
or unlocking other functionality, check out
[Fairwinds Insights](https://fairwinds.com/insights),
a platform for auditing and enforcing policy in Kubernetes clusters.
