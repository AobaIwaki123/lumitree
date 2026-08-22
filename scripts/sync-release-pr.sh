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
PREV_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v1.0.0")
CLEAN_VER=$(echo "$PREV_TAG" | sed -E 's/^(lumitree-)?v?//')
MAJOR=$(echo "$CLEAN_VER" | cut -d. -f1)
MINOR=$(echo "$CLEAN_VER" | cut -d. -f2)
NEXT_MINOR=$((MINOR + 1))
NEXT_TAG="v${MAJOR}.${NEXT_MINOR}.0"

echo "Updating Kubernetes manifests to next release version ${NEXT_TAG}..."
sed -i.bak -E "s|(image: ghcr\.io/aobaiwaki123/lumitree:).*|\1${NEXT_TAG}|" k8s/manifests/deployment.yml
sed -i.bak -E "s/(newTag: ).*/\1${NEXT_TAG}/" k8s/manifests/kustomization.yml
rm -f k8s/manifests/*.bak

git config user.name "github-actions[bot]" 2>/dev/null || true
git config user.email "github-actions[bot]@users.noreply.github.com" 2>/dev/null || true

git add k8s/manifests/deployment.yml k8s/manifests/kustomization.yml
if ! git diff --cached --quiet; then
  git commit -m "chore(release): bump k8s manifest image tag to ${NEXT_TAG}"
  git push origin main
  echo "Successfully committed and pushed bumped k8s manifests to main."
fi

PR_TITLE="release: 本番リリース ${NEXT_TAG} (${DATE})"

PR_BODY=$(cat <<EOF
## 本番リリース PR: \`main\` -> \`release\` (${NEXT_TAG})

本 PR は \`main\` ブランチの開発成果を本番 \`release\` ブランチへ反映し、新バージョン **${NEXT_TAG}** をリリースするための PR です。

### 変更・コミット一覧
\`\`\`
${UNRELEASED_COMMITS}
\`\`\`

### リリース後の自動実行項目
- [ ] 次期 Git リリースタグ (**${NEXT_TAG}**) の自動発行
- [ ] GitHub Release ノートの自動生成 & GoReleaser バイナリ配布
- [ ] GHCR へのマルチアーキテクチャ OCI コンテナイメージ (\`ghcr.io/aobaiwaki123/lumitree:${NEXT_TAG}\`) 自動 Push
- [ ] 自宅 Kubernetes クラスタ (ArgoCD) への自動同期 & ローリングアップデート
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
