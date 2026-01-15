#!/bin/bash

set -e

VERSION=$1

if [ -z "$VERSION" ]; then
    echo "❌ Usage: ./scripts/release.sh <version>"
    echo "Example: ./scripts/release.sh 0.2.1"
    exit 1
fi

TAG="v${VERSION}"

echo "🚀 Creating release ${TAG}..."
echo ""

# Check if tag already exists
if git rev-parse "$TAG" >/dev/null 2>&1; then
    echo "❌ Tag ${TAG} already exists!"
    echo "To update, delete it first:"
    echo "  git tag -d ${TAG}"
    echo "  git push origin --delete ${TAG}"
    exit 1
fi

# Check if working directory is clean
if ! git diff-index --quiet HEAD --; then
    echo "⚠️  Working directory has uncommitted changes!"
    echo ""
    echo "Please commit or stash your changes first:"
    echo "  git add ."
    echo "  git commit -m 'chore: prepare release ${TAG}'"
    echo ""
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Show what will be tagged
echo "📋 Current commit:"
git log -1 --oneline
echo ""

# Confirm
read -p "Create tag ${TAG} at this commit? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelled."
    exit 1
fi

# Create annotated tag
echo "🏷️  Creating annotated tag..."
git tag -a "$TAG" -m "Release ${TAG}

Features:
- Default ERC20 rule untuk auto-detect semua Transfer/Mint/Burn events
- Standard ERC20 ABI terintegrasi
- Auto-decode events menggunakan ABI registry
- Console logging untuk demo

See CHANGELOG.md for full details."

# Push tag
echo "📤 Pushing tag to remote..."
git push origin "$TAG"

echo ""
echo "✅ Release ${TAG} created successfully!"
echo ""
echo "📝 Next steps:"
echo "1. Create GitHub release:"
echo "   gh release create ${TAG} --title 'v${VERSION} - Default ERC20 Rule & Auto Decode' --notes-file CHANGELOG.md"
echo ""
echo "2. Or create release via GitHub web UI:"
echo "   https://github.com/YOUR_USERNAME/asentric-sdk/releases/new"
echo ""
echo "3. Update dependencies in other projects:"
echo "   go get github.com/asentric/asentric@${TAG}"
echo ""
