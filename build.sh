#!/usr/bin/env bash
set -euo pipefail

mkdir -p build
go build -o build/loadsim .
echo "built: build/loadsim"
