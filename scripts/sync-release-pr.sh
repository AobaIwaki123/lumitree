#!/usr/bin/env bash
# ==============================================================================
# sync-release-pr.sh - Automated Release PR creator/updater (main -> release)
# ==============================================================================

set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_DIR"

echo "Fetching latest branches and tags..."
git fetch --tags origin main release || true

UNRELEASED_COMMITS=$(git log origin/release..origin/main --oneline --no-merges || true)

if [ -z "$UNRELEASED_COMMITS" ]; then
  echo "No unreleased commits found between origin/release and origin/main. Nothing to do."
  exit 0
fi

echo "Found unreleased commits:"
echo "$UNRELEASED_COMMITS"
echo ""

# Extract merged PR references (#XX) from merge commits (GitHub native rich preview)
MERGED_PRS=$(git log origin/release..origin/main --merges --oneline | grep -oE '#[0-9]+' | sort -V -r -u || true)
RELEASE_NOTES_ITEMS=""
if [ -n "$MERGED_PRS" ]; then
  while read -r pr_ref; do
    if [ -n "$pr_ref" ]; then
      RELEASE_NOTES_ITEMS="${RELEASE_NOTES_ITEMS}
- ${pr_ref}"
    fi
  done <<< "$MERGED_PRS"
fi

if [ -z "$RELEASE_NOTES_ITEMS" ]; then
  RELEASE_NOTES_ITEMS=$(echo "$UNRELEASED_COMMITS" | sed 's/^/- /')
fi

DATE=$(date +%Y-%m-%d)
PREV_TAG=$(git tag -l 'v*' | sort -V | tail -n 1)
if [ -z "$PREV_TAG" ]; then
  PREV_TAG="v1.0.0"
fi
echo "Current highest version tag: ${PREV_TAG}"
CLEAN_VER=$(echo "$PREV_TAG" | sed -E 's/^(lumitree-)?v?//')
MAJOR=$(echo "$CLEAN_VER" | cut -d. -f1)
MINOR=$(echo "$CLEAN_VER" | cut -d. -f2)
NEXT_MINOR=$((MINOR + 1))
NEXT_TAG="v${MAJOR}.${NEXT_MINOR}.0"
echo "Calculated next release tag: ${NEXT_TAG}"

STAGE_BRANCH="release-stage/${NEXT_TAG}"
echo "Preparing staging branch: ${STAGE_BRANCH}..."

# Create/reset staging branch from origin/release (ensures fast-forward & zero manifest conflict)
git checkout -B "$STAGE_BRANCH" origin/release

# Merge all development changes from origin/main into staging branch
git merge -X theirs origin/main -m "chore: sync main into ${STAGE_BRANCH}" 2>/dev/null || true

echo "Updating Kubernetes manifests to release version ${NEXT_TAG} on ${STAGE_BRANCH}..."
sed -i.bak -E "s|(image: ghcr\.io/aobaiwaki123/lumitree:).*|\1${NEXT_TAG}|" k8s/manifests/deployment.yml
sed -i.bak -E "s/(newTag: ).*/\1${NEXT_TAG}/" k8s/manifests/kustomization.yml
rm -f k8s/manifests/*.bak

git config user.name "github-actions[bot]" 2>/dev/null || true
git config user.email "github-actions[bot]@users.noreply.github.com" 2>/dev/null || true

git add k8s/manifests/deployment.yml k8s/manifests/kustomization.yml
if ! git diff --cached --quiet; then
  git commit -m "chore(release): bump k8s manifest image tag to ${NEXT_TAG}"
fi

echo "Pushing staging branch ${STAGE_BRANCH} to origin..."
git push -u --force origin "$STAGE_BRANCH"

PR_TITLE="release: 本番リリース ${NEXT_TAG} (${DATE})"

PR_BODY=$(cat <<EOF
## 本番リリース PR: \`${STAGE_BRANCH}\` -> \`release\` (${NEXT_TAG})

本 PR は \`main\` ブランチの開発成果を取りまとめ、マニフェストタグを **${NEXT_TAG}** に更新して本番 \`release\` ブランチへ反映するための Release PR です。

### 📦 含まれる変更・機能一覧
${RELEASE_NOTES_ITEMS}

### リリース後の自動実行項目
- [ ] 次期 Git リリースタグ (**${NEXT_TAG}**) の自動発行
- [ ] GitHub Release ノートの自動生成 & GoReleaser バイナリ配布
- [ ] GHCR へのマルチアーキテクチャ OCI コンテナイメージ (\`ghcr.io/aobaiwaki123/lumitree:${NEXT_TAG}\`) 自動 Push
- [ ] 自宅 Kubernetes クラスタ (ArgoCD) への自動同期 & ローリングアップデート
- [ ] 一時ステージングブランチ (\`${STAGE_BRANCH}\`) の自動削除
EOF
)

EXISTING_PR=$(gh pr list --base release --head "$STAGE_BRANCH" --json number --jq '.[0].number' || true)

if [ -n "$EXISTING_PR" ]; then
  echo "Updating existing Release PR #${EXISTING_PR}..."
  gh pr edit "$EXISTING_PR" --title "$PR_TITLE" --body "$PR_BODY"
  echo "Successfully updated Release PR #${EXISTING_PR}."
else
  echo "Creating new Release PR from ${STAGE_BRANCH} to release..."
  gh pr create \
    --base release \
    --head "$STAGE_BRANCH" \
    --title "$PR_TITLE" \
    --body "$PR_BODY"
  echo "Successfully created Release PR."
fi
