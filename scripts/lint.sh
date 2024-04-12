#!/usr/bin/env bash
set -euo pipefail
IFS=$' \t\n'

glciVersion="v1.57.2"

cd "$(dirname "${0}")"/..

result=$(go fmt ./...)
if [[ ${result} ]]; then
	echo "The following files are not formatted according to 'go fmt'"
	echo "${result}"
	exit 1
fi

if [ "${1-}" == "fix" ]; then
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@"${glciVersion}" run --fix
else
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@"${glciVersion}" run
fi

shopt -s globstar nullglob
for file in ./scripts/**; do
	if [[ "${file}" =~ .*\.sh$ ]]; then
		shellcheck --severity=info "${file}"
	fi
done

for file in ./proxy/**; do
	if [[ "${file}" =~ .*\.sh$ ]]; then
		shellcheck --severity=info "${file}"
	fi
done
