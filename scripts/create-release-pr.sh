#!/usr/bin/env bash
# ==============================================================================
# create-release-pr.sh - Release Branch & PR Automation Script for lumitree
#
# Usage:
#   ./scripts/create-release-pr.sh <version>
# Example:
#   ./scripts/create-release-pr.sh 1.1.0
# ==============================================================================

set -euo pipefail

VERSION="${1:-}"

if [[ -z "$VERSION" ]]; then
  echo "❌ Error: Version is required."
  echo "Usage: $0 <version> (e.g. 1.1.0)"
  exit 1
fi

# Remove leading 'v' if present
VERSION="${VERSION#v}"
TAG="v${VERSION}"
BRANCH="release/${TAG}"

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_DIR"

echo "🚀 Preparing release ${TAG} on branch ${BRANCH}..."

# Ensure working directory is clean
if [[ -n "$(git status --porcelain)" ]]; then
  echo "❌ Error: Working directory has uncommitted changes. Please commit or stash them."
  exit 1
fi

# Fetch and checkout latest main
git fetch origin main
git checkout main
git pull origin main

# Create release branch
if git show-ref --verify --quiet "refs/heads/${BRANCH}"; then
  echo "⚠️ Local branch ${BRANCH} already exists. Checking it out..."
  git checkout "${BRANCH}"
else
  git checkout -b "${BRANCH}" main
fi

# 1. Update version variable in cmd/lumitree/main.go
echo "📝 Updating version in cmd/lumitree/main.go..."
sed -i '' "s/version = \".*\"/version = \"${VERSION}\"/" cmd/lumitree/main.go

# 2. Generate / Update CHANGELOG.md
echo "📋 Generating changelog..."
PREV_TAG="$(git describe --tags --abbrev=0 2>/dev/null || echo "")"
DATE="$(date +%Y-%m-%d)"

CHANGELOG_ENTRY="## [${TAG}] - ${DATE}"
if [[ -n "$PREV_TAG" ]]; then
  COMMITS="$(git log "${PREV_TAG}..HEAD" --oneline --no-merges || true)"
else
  COMMITS="$(git log --oneline --no-merges -n 20 || true)"
fi

if [[ -f "CHANGELOG.md" ]]; then
  TMP_CHANGELOG="$(mktemp)"
  {
    echo "$CHANGELOG_ENTRY"
    echo ""
    echo "$COMMITS" | sed 's/^/- /'
    echo ""
    cat CHANGELOG.md
  } > "$TMP_CHANGELOG"
  mv "$TMP_CHANGELOG" CHANGELOG.md
else
  {
    echo "# Changelog"
    echo ""
    echo "$CHANGELOG_ENTRY"
    echo ""
    echo "$COMMITS" | sed 's/^/- /'
  } > CHANGELOG.md
fi

# Run tests and verify
echo "🧪 Running tests..."
go test ./...

# Commit changes
git add cmd/lumitree/main.go CHANGELOG.md
git commit -m "chore(release): prepare ${TAG}"

# Push and create PR
echo "📤 Pushing branch to origin..."
git push -u origin "${BRANCH}"

echo "📬 Creating Release Pull Request..."
gh pr create \
  --base main \
  --head "${BRANCH}" \
  --title "release: ${TAG}" \
  --body "## Release ${TAG}

This PR prepares the release for **${TAG}**.

### Changes in this Release
\`\`\`
${COMMITS}
\`\`\`

### Release Checklist
- [ ] Version in \`cmd/lumitree/main.go\` is updated to \`${VERSION}\`
- [ ] \`CHANGELOG.md\` is reviewed and accurate
- [ ] All CI checks pass
- [ ] When this PR is merged to \`main\`, the release tag \`${TAG}\` will be created automatically"

echo "✅ Release PR created successfully!"
