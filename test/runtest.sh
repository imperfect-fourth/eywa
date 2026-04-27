#!/usr/bin/env bash

set -euo pipefail

go run ../cmd/eywagen/main.go -types User,UserSubField,NotUser,arguments
diff eywa_generated.go eywa_generated.go.golden || echo "ERROR: golden file differs" && exit 1

go test .
