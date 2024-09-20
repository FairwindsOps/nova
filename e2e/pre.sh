#!/bin/bash

set -e

make build-linux
docker cp ./ e2e-command-runner:/nova