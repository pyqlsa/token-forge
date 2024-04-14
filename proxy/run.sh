#!/usr/bin/env bash
set -euo pipefail
IFS=$' \t\n'

cd "$(dirname "${0}")"/..

#docker run -it --rm --name proxy -p 9050:9050 proxy:latest
docker run -it --rm --name proxy -p 127.0.0.1:9050:9050/tcp -p 127.0.0.1:9080:9080/tcp -p 127.0.0.1:8853:8853/udp proxy:latest
