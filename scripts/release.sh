#!/bin/bash
set -e

# Get current version
CURRENT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")

# Parse version parts
VERSION_NO_V=${CURRENT_VERSION#v}
IFS='.' read -r MAJOR MINOR PATCH <<< "$VERSION_NO_V"

# Determine new version based on bump type
case "$1" in
  major)
    NEW_VERSION="v$((MAJOR + 1)).0.0"
    ;;
  minor)
    NEW_VERSION="v${MAJOR}.$((MINOR + 1)).0"
    ;;
  patch)
    NEW_VERSION="v${MAJOR}.${MINOR}.$((PATCH + 1))"
    ;;
  *)
    echo "Usage: $0 {major|minor|patch}"
    exit 1
    ;;
esac

# Display version info
echo "Current version: $CURRENT_VERSION"
echo "New version:     $NEW_VERSION"
echo ""

# Confirm with user
read -p "Create and push tag $NEW_VERSION? [y/N] " -n 1 -r
echo ""

if [[ $REPLY =~ ^[Yy]$ ]]; then
  git tag -s "$NEW_VERSION" -m "Release $NEW_VERSION"
  git push origin "$NEW_VERSION"
  echo "✅ Released $NEW_VERSION"
else
  echo "❌ Release cancelled"
  exit 1
fi

