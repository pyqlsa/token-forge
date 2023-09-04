#!/usr/bin/env bash
set -euo pipefail
IFS=$' \t\n'

cd "$(dirname "${0}")"/..

version=$(git describe --tags --always)
platforms=("linux" "darwin" "windows")
arches=("amd64" "arm64")
pkgs=("token-forge")

ldflags="-X 'main.version=${version}' -X 'main.commit=$(git rev-parse HEAD)'"
for platform in "${platforms[@]}"; do
  for arch in "${arches[@]}"; do
    [[ "${platform}" == "windows" && "${arch}" == "arm64" ]] && continue # skip it
    for pkg in "${pkgs[@]}"; do
      echo "---------- building ${platform}-${arch}/${pkg} ----------"
      mkdir -p ./build/"${platform}-${arch}"/
      GOOS="${platform}" GOARCH="${arch}" go build -trimpath -v -ldflags "${ldflags}" -o "./build/${platform}-${arch}/${pkg}" "./cmd/${pkg}"
    done
  done
done
