---
meta:
  - name: description
    content: "Fairwinds Nova | Contribution Guidelines"
---
# Contributing

Issues, whether bugs, tasks, or feature requests are essential for keeping our projects great.
We believe it should be as easy as possible to contribute changes that get things working in your environment.
There are a few guidelines that we need contributors to follow so that we can keep on top of things.

## Code of Conduct

This project adheres to a [code of conduct](/contributing/code-of-conduct). Please review this document before contributing to this project.

## Sign the CLA
Before you can contribute, you will need to sign the [Contributor License Agreement](https://cla-assistant.io/fairwindsops/nova).

## Creating a New Issue

If you've encountered an issue that is not already reported, please create a [new issue](https://github.com/FairwindsOps/nova/issues), choose `Bug Report`, `Feature Request` or `Misc.` and follow the instructions in the template. 

## Getting Started

We label issues with the ["good first issue" tag](https://github.com/FairwindsOps/nova/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22)
if we believe they'll be a good starting point for new contributors. If you're interested in working on an issue,
please start a conversation on that issue, and we can help answer any questions as they come up.

## Setting Up Your Development Environment
### Prerequisites
* A properly configured Golang environment with Go 1.11 or higher
* A Kubernetes cluster defined in `~/.kube/config` (or in a file specified by the `KUBECONFIG` env variable)

### Installation
* Install the project with `go get github.com/fairwindsops/nova`
* Change into the Nova directory which is installed at `$GOPATH/src/github.com/fairwindsops/nova`
* See the results with `go run main.go find`

## Running Tests

The following commands are all required to pass as part of testing:

```
go list ./... | grep -v vendor | xargs golint -set_exit_status
go list ./... | grep -v vendor | xargs go vet
go test ./...
```

## Creating a Pull Request

Each new pull request should:

- Reference any related issues
- Add tests that show the issues have been solved
- Pass existing tests and linting
- Contain a clear indication of if they're ready for review or a work in progress
- Be up to date and/or rebased on the master branch

## Creating a new release
* Update version in README.md
* Update version in main.go
* Open and merge a PR
* Tag your merged commit

Tagging a commit will build and push a new Docker image, add a binary to the
[releases page](https://github.com/FairwindsOps/nova/releases), and publish a new Homebrew binary.

Be sure to add any notes to CHANGELOG.md

