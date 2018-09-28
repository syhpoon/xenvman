#!/bin/bash

set -e

dirs=$(find . -maxdepth 1 -type d -and -not -path '*/\.*' | grep -Ev 'vendor|etc|tmp|docs' | grep -v "^.$")

for dir in ${dirs}; do
   echo "*** Formatting $dir"
   go fmt "$dir/..."

   echo "*** Testing $dir"
   go test -cover "$dir/..." || exit

   echo "*** Vetting $dir"
   go vet "$dir/..."
done

## Run linter
glint_bin=$(command -v golangci-lint || true)

if [ -z "${glint_bin}" ]; then
   echo "!!! please consider installing github.com/golangci/golangci-lint"
else
	${glint_bin} run --enable-all -e etc -e vendor
fi
