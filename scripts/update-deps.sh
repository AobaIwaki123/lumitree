#!/usr/bin/env bash
# ==============================================================================
# lumitree Dependency & CI Version Updater
#
# Checks and updates:
#   - Go module dependencies and Go version
#   - GitHub Actions workflow versions (actions/*, goreleaser, golangci-lint)
#   - Dockerfile base images
#   - Runs test suite to verify compatibility
# ==============================================================================

set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_DIR"

echo "🔍 Checking and updating dependencies in $REPO_DIR..."

# 1. Update Go dependencies and tidy
echo "📦 Updating Go dependencies..."
go mod tidy

# 2. Recommended Versions
TARGET_GO_VERSION="1.23"
GOLANGCI_LINT_ACTION_VER="v6"
GORELEASER_ACTION_VER="v6"

echo "⚙️ Synchronizing CI and Dockerfile configurations..."

safe_sed() {
  local pattern="$1"
  local file="$2"
  if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' "$pattern" "$file"
  else
    sed -i "$pattern" "$file"
  fi
}

# Update CI Go Workflow
if [[ -f ".github/workflows/ci-go.yml" ]]; then
  safe_sed "s/go-version: .*/go-version: '${TARGET_GO_VERSION}'/" .github/workflows/ci-go.yml
  safe_sed "s|uses: golangci/golangci-lint-action@.*|uses: golangci/golangci-lint-action@${GOLANGCI_LINT_ACTION_VER}|" .github/workflows/ci-go.yml
fi

# Update Release Workflow
if [[ -f ".github/workflows/release.yml" ]]; then
  safe_sed "s/go-version: .*/go-version: '${TARGET_GO_VERSION}'/" .github/workflows/release.yml
  safe_sed "s|uses: goreleaser/goreleaser-action@.*|uses: goreleaser/goreleaser-action@${GORELEASER_ACTION_VER}|" .github/workflows/release.yml
fi

# Update Live Monitor Workflow
if [[ -f ".github/workflows/live-monitor.yml" ]]; then
  safe_sed "s/go-version: .*/go-version: '${TARGET_GO_VERSION}'/" .github/workflows/live-monitor.yml
fi

# Update Dockerfile
if [[ -f "Dockerfile" ]]; then
  safe_sed "s|FROM golang:.* AS builder|FROM golang:${TARGET_GO_VERSION}-alpine AS builder|" Dockerfile
fi

echo "🧪 Running tests to verify changes..."
go test -v -race ./...
go build -o bin/lumitree ./cmd/lumitree

echo "✅ All checks passed! Versions updated to Go ${TARGET_GO_VERSION}."
