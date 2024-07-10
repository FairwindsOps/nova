#!/bin/bash

set -e

make build
docker cp ./ e2e-command-runner:/pluto