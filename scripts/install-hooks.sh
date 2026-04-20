#!/bin/sh

set -eu

HOOK_DIR=$(git rev-parse --git-dir)/hooks
SCRIPT_DIR=$(CDPATH= cd -- "$(dirname "$0")" && pwd)

echo "Installing git hooks..."

echo "Installing pre-commit hook..."
ln -sf "$SCRIPT_DIR/git-hooks/pre-commit" "$HOOK_DIR/pre-commit"

echo "Installing pre-push hook..."
ln -sf "$SCRIPT_DIR/git-hooks/pre-push" "$HOOK_DIR/pre-push"

chmod +x "$SCRIPT_DIR/git-hooks/pre-commit" "$SCRIPT_DIR/git-hooks/pre-push"

echo "Git hooks installed successfully"
