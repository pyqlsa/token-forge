#!/usr/bin/env bash
set -euo pipefail
IFS=$' \t\n'

cd "$(dirname "${0}")"/..

if [ "${1-}" == "race" ]; then
  go test -v -failfast -race ./...
else
  go test -v -failfast ./...
fi

