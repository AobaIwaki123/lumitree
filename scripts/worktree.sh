#!/usr/bin/env bash
# ==============================================================================
# worktree.sh - Git Worktree management helper script for lumitree
#
# Usage:
#   ./scripts/worktree.sh create <branch-name> [base-branch]
#   ./scripts/worktree.sh list
#   ./scripts/worktree.sh remove <branch-name>
#   ./scripts/worktree.sh clean
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
    git worktree remove "$TARGET_DIR" --force 2>/dev/null || true
    rm -rf "$TARGET_DIR"
    git worktree prune

    if git show-ref --verify --quiet "refs/heads/${BRANCH_NAME}"; then
      echo "Deleting local branch: ${BRANCH_NAME}..."
      git branch -D "$BRANCH_NAME" || true
    fi

    echo "Worktree and local branch '${BRANCH_NAME}' successfully removed."
    ;;

  clean)
    echo "Pruning remote-tracking branches..."
    git fetch --prune origin || true

    echo "Scanning for merged or gone worktrees and local branches..."
    git worktree list --porcelain | grep "^worktree " | cut -d' ' -f2- | while read -r wt_path; do
      if [ "$wt_path" != "$REPO_DIR" ] && [ -d "$wt_path" ]; then
        wt_branch=$(git -C "$wt_path" branch --show-current 2>/dev/null || true)
        if [ -n "$wt_branch" ]; then
          # Check if merged to main or remote branch is gone
          if git branch -r | grep -q "origin/${wt_branch}"; then
            # Remote exists, check if merged to main
            if git log "origin/main..${wt_branch}" --oneline 2>/dev/null | grep -q .; then
              continue # Has unmerged commits, keep
            fi
          fi
          echo "Cleaning up merged worktree: ${wt_path} (Branch: ${wt_branch})..."
          git worktree remove "$wt_path" --force 2>/dev/null || true
          rm -rf "$wt_path"
          git branch -D "$wt_branch" 2>/dev/null || true
        fi
      fi
    done
    git worktree prune

    # Delete any gone local branches without worktrees
    git branch -vv | grep ': gone]' | awk '{print $1}' | tr -d '*+' | while read -r gone_branch; do
      if [ -n "$gone_branch" ] && [ "$gone_branch" != "main" ] && [ "$gone_branch" != "release" ]; then
        echo "Deleting orphaned gone local branch: ${gone_branch}..."
        git branch -D "$gone_branch" 2>/dev/null || true
      fi
    done

    echo "Clean complete. Remaining worktrees:"
    git worktree list
    ;;

  *)
    echo "Usage: $0 {create|list|remove|clean} [branch-name] [base-branch]"
    exit 1
    ;;
esac
