#!/usr/bin/env bash
# ==============================================================================
# worktree.sh - Git Worktree management helper script for lumitree
#
# Usage:
#   ./scripts/worktree.sh create <branch-name> [base-branch]
#   ./scripts/worktree.sh list
#   ./scripts/worktree.sh remove <branch-name>
# ==============================================================================

set -euo pipefail

ACTION="${1:-list}"
BRANCH_NAME="${2:-}"
BASE_BRANCH="${3:-main}"

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORKTREE_BASE_DIR="${REPO_DIR}/.worktrees"

case "$ACTION" in
  create)
    if [ -z "$BRANCH_NAME" ]; then
      echo "Error: Branch name is required."
      echo "Usage: $0 create <branch-name> [base-branch]"
      exit 1
    fi

    TARGET_DIR="${WORKTREE_BASE_DIR}/${BRANCH_NAME}"
    echo "Creating Git Worktree at: ${TARGET_DIR} (Base: ${BASE_BRANCH})..."
    mkdir -p "$(dirname "$TARGET_DIR")"

    git fetch origin "$BASE_BRANCH" || true

    if git show-ref --verify --quiet "refs/heads/${BRANCH_NAME}"; then
      git worktree add "$TARGET_DIR" "$BRANCH_NAME"
    else
      git worktree add -b "$BRANCH_NAME" "$TARGET_DIR" "origin/${BASE_BRANCH}"
    fi

    echo "Worktree ready at: ${TARGET_DIR}"
    ;;

  list)
    echo "Current Git Worktrees:"
    git worktree list
    ;;

  remove)
    if [ -z "$BRANCH_NAME" ]; then
      echo "Error: Branch name is required."
      echo "Usage: $0 remove <branch-name>"
      exit 1
    fi

    TARGET_DIR="${WORKTREE_BASE_DIR}/${BRANCH_NAME}"
    echo "Removing Git Worktree: ${TARGET_DIR}..."
    git worktree remove "$TARGET_DIR" --force || true
    rm -rf "$TARGET_DIR"
    git worktree prune
    echo "Worktree removed."
    ;;

  *)
    echo "Usage: $0 {create|list|remove} [branch-name] [base-branch]"
    exit 1
    ;;
esac
