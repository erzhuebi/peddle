#!/bin/sh
set -eu

DOCS_DIR="docs/language"
WIKI_REPO="git@github.com:erzhuebi/peddle.wiki.git"
WIKI_BRANCH="master"

SSH_KEY_NAME="${1:-}"
SSH_KEY=""

if [ ! -d "$DOCS_DIR" ]; then
    echo "Docs directory not found: $DOCS_DIR"
    echo "Run this script from the peddle repository root."
    exit 1
fi

if [ -n "$SSH_KEY_NAME" ]; then
    case "$SSH_KEY_NAME" in
        /*)
            SSH_KEY="$SSH_KEY_NAME"
            ;;
        *)
            SSH_KEY="$HOME/.ssh/$SSH_KEY_NAME"
            ;;
    esac

    if [ ! -f "$SSH_KEY" ]; then
        echo "SSH key not found: $SSH_KEY"
        exit 1
    fi

    GIT_SSH_COMMAND="ssh -i $SSH_KEY -o IdentitiesOnly=yes"
    export GIT_SSH_COMMAND
fi

TMP_DIR="$(mktemp -d)"
cleanup() {
    rm -rf "$TMP_DIR"
}
trap cleanup EXIT INT TERM

WIKI_DIR="$TMP_DIR/peddle.wiki"

echo "Cloning wiki..."
git clone "$WIKI_REPO" "$WIKI_DIR"

cd "$WIKI_DIR"

if git show-ref --verify --quiet "refs/heads/$WIKI_BRANCH"; then
    git checkout "$WIKI_BRANCH"
elif git show-ref --verify --quiet "refs/remotes/origin/$WIKI_BRANCH"; then
    git checkout -b "$WIKI_BRANCH" "origin/$WIKI_BRANCH"
else
    echo "Wiki branch '$WIKI_BRANCH' not found, using current branch."
fi

cd - >/dev/null

echo "Copying documentation..."
cp "$DOCS_DIR"/*.md "$WIKI_DIR"/

cd "$WIKI_DIR"

git add .

if git diff --cached --quiet; then
    echo "No wiki documentation changes to commit."
    exit 0
fi

git commit -m "Sync language documentation"

echo "Pushing wiki..."
git push

echo "Wiki documentation published."
