#!/bin/bash

# Git hooksを.git/hooksディレクトリにシンボリックリンクとして作成

HOOK_DIR=$(git rev-parse --git-dir)/hooks
SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)

echo "Installing git hooks..."

# pre-commitフックのインストール
echo "Installing pre-commit hook..."
ln -sf "$SCRIPT_DIR/git-hooks/pre-commit" "$HOOK_DIR/pre-commit"

echo "Git hooks installed successfully"