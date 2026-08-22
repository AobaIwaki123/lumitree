#!/usr/bin/env bash
# ==============================================================================
# sync-release-pr.sh - Automated Release PR creator/updater (main -> release)
# ==============================================================================

set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_DIR"

echo "Fetching latest branches..."
git fetch origin main release || true

UNRELEASED_COMMITS=$(git log origin/release..origin/main --oneline --no-merges || true)

if [ -z "$UNRELEASED_COMMITS" ]; then
  echo "No unreleased commits found between origin/release and origin/main. Nothing to do."
  exit 0
fi

echo "Found unreleased commits:"
echo "$UNRELEASED_COMMITS"
echo ""

DATE=$(date +%Y-%m-%d)
PR_TITLE="release: sync main to release (${DATE})"

PR_BODY=$(cat <<EOF
## Release PR: \`main\` -> \`release\`

This PR bundles all changes from \`main\` ready to be promoted to the production \`release\` branch.

### Unreleased Commits:
\`\`\`
${UNRELEASED_COMMITS}
\`\`\`

### Checklist for Release:
- [ ] Review all changes above
- [ ] CI checks on this PR are 100% Green
- [ ] Merging this PR into \`release\` will trigger production container build and release tagging.
EOF
)

EXISTING_PR=$(gh pr list --base release --head main --json number --jq '.[0].number' || true)

if [ -n "$EXISTING_PR" ]; then
  echo "Updating existing Release PR #${EXISTING_PR}..."
  gh pr edit "$EXISTING_PR" --title "$PR_TITLE" --body "$PR_BODY"
  echo "Successfully updated Release PR #${EXISTING_PR}."
else
  echo "Creating new Release PR from main to release..."
  gh pr create \
    --base release \
    --head main \
    --title "$PR_TITLE" \
    --body "$PR_BODY"
  echo "Successfully created Release PR."
fi
