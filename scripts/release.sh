#!/bin/bash
set -e

# Refuse to tag a dirty tree or a stale main.
if [ -n "$(git status --porcelain)" ]; then
  echo "Working tree is dirty. Commit or stash first." >&2
  exit 1
fi

BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [ "$BRANCH" != "main" ]; then
  echo "Not on main (current: $BRANCH). Release tags must be cut from main." >&2
  exit 1
fi

git fetch origin main --quiet
LOCAL=$(git rev-parse main)
REMOTE=$(git rev-parse origin/main)
if [ "$LOCAL" != "$REMOTE" ]; then
  echo "Local main is not in sync with origin/main. Pull or push first." >&2
  exit 1
fi

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
  echo "Released $NEW_VERSION"
else
  echo "Release cancelled"
  exit 1
fi

