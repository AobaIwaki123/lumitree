#!/usr/bin/env bash
# ==============================================================================
# verify-all.sh - Strict local pre-push verification script for lumitree
#
# This script executes all CI checks locally to guarantee that code will pass CI
# before pushing any commits or creating PRs.
# ==============================================================================

set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_DIR"

echo "========================================================"
echo "1. Checking Schema Drift (go generate & git diff)..."
echo "========================================================"
go generate ./...
if ! git diff --exit-code; then
  echo "Error: Uncommitted generated code detected! Please commit generated files."
  exit 1
fi
echo "OK: Code generation is up to date."
echo ""

echo "========================================================"
echo "2. Running golangci-lint..."
echo "========================================================"
GOPATH_BIN="$(go env GOPATH)/bin"
if [[ -f "${GOPATH_BIN}/golangci-lint" ]]; then
  "${GOPATH_BIN}/golangci-lint" run ./...
else
  golangci-lint run ./...
fi
echo "OK: Linter passed with 0 issues."
echo ""

echo "========================================================"
echo "3. Running Unit, Mock & Integration Tests (-race)..."
echo "========================================================"
go test -race -v -cover ./...
echo "OK: All tests passed."
echo ""

echo "========================================================"
echo "4. Building all binaries (cmd/lumitree)..."
echo "========================================================"
mkdir -p bin
go build -v -o bin/lumitree ./cmd/lumitree
echo "OK: Binary build successful."
echo ""

echo "========================================================"
echo "All local verification checks passed with 100% success!"
echo "========================================================"
