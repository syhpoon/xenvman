#!/bin/sh

set -e

cd pkg/server/webapp
npm run build
cd ..
go generate
