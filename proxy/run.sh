#!/usr/bin/env bash
set -euo pipefail
IFS=$' \t\n'

cd "$(dirname "${0}")"/..

#docker run -it --rm --name proxy -p 9050:9050 proxy:latest
docker run -it --rm --name proxy -p 127.0.0.1:9050:9050/tcp proxy:latest

