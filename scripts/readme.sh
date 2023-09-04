#!/usr/bin/env bash
set -euo pipefail
IFS=$' \t\n'

cd "$(dirname "${0}")"/..

./scripts/build.sh

go run ./cmd/readme/main.go
