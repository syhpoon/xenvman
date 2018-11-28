#!/bin/bash
# Usage: go-tool.sh <tool> <Message>

set -e

tool=$1
msg=$2

dirs=$(find . -maxdepth 1 -type d -and -not -path '*/\.*' | grep -Ev 'vendor|etc|tmp|docs|mk|tpl|install' | grep -v "^.$")

for dir in ${dirs}; do
   echo "*** ${msg} ${dir}";
   sh -c "${tool} ./${dir}/...";
done
