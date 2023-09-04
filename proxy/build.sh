#!/usr/bin/env bash
set -euo pipefail
IFS=$' \t\n'

cd "$(dirname "${0}")"/..

docker build -t proxy -f ./proxy/Dockerfile .

