#!/bin/bash

# Configuration
REPO_URL="https://github.com/pixelvide/cloud-sentinel-k8s.git"
# Fetch all tags
git fetch --tags

# Get all semantic version tags (v*), sort them (version sort), and take the last 5 relevant ones
# You can adjust the grep/sort logic to fit your tagging scheme
VERSIONS=($(git tag -l "v*" | grep -v "v0.0.0" | sort -V | tail -n 20))
echo "Detected versions to build: ${VERSIONS[*]}"
BASE_URL="/cloud-sentinel-k8s"

# Setup directories
WORK_DIR=$(pwd)
DOCS_DIR="$WORK_DIR/docs"
DIST_DIR="$DOCS_DIR/.vitepress/dist"
TEMP_DIR=$(mktemp -d)

echo "Cleaning release directory..."
rm -rf "$DIST_DIR"

# 1. Build Current Version (latest)
echo "Building 'latest' documentation..."
cd "$DOCS_DIR"
# Ensure the base URL is correct for the main build
# This assumes the config.mts has base: '/cloud-sentinel-k8s/'
pnpm install
pnpm run docs:build
cd "$WORK_DIR"

# 2. Build Previous Versions
# We need to clone the repo to a temp location to avoid messing up the current workspace
echo "Cloning repository for historical versions..."
git clone "$REPO_URL" "$TEMP_DIR/repo"

for VERSION in "${VERSIONS[@]}"; do
  echo "------------------------------------------------"
  echo "Building version: $VERSION"
  
  VERSION_DIR="$TEMP_DIR/repo-$VERSION"
  cp -r "$TEMP_DIR/repo" "$VERSION_DIR"
  cd "$VERSION_DIR"
  
  # Checkout the specific tag
  git checkout "$VERSION"
  
  # Check if docs directory exists in this version
  if [ ! -d "docs" ]; then
    echo "Warning: docs directory not found in $VERSION. Skipping."
    continue
  fi
  
  cd docs
  
  # Install dependencies for that old version
  # We might need to use 'npm' or 'pnpm' depending on what was used then
  # Assuming pnpm for consistency, but falling back might be needed
  if [ -f "pnpm-lock.yaml" ]; then
    pnpm install
  else
    npm install
  fi
  
  # IMPORTANT: We must dynamically inject the correct base URL for this version
  # We use sed to replace the base URL in the config file
  # Pattern matches: base: "..." or base: '...'
  # We replace it with: base: "/cloud-sentinel-k8s/$VERSION/"
  
  CONFIG_FILE=".vitepress/config.mts"
  if [ ! -f "$CONFIG_FILE" ]; then
    CONFIG_FILE=".vitepress/config.ts"
  fi
  if [ ! -f "$CONFIG_FILE" ]; then
      CONFIG_FILE=".vitepress/config.js"
  fi
  
  # This sed command is a bit fragile but works for standard configs
  # It looks for "base: ..." and replaces it.
  if [[ "$OSTYPE" == "darwin"* ]]; then
      # MacOS sed requires empty string for backup extension
      sed -i '' "s|base: \".*\"|base: \"$BASE_URL/$VERSION/\"|g" "$CONFIG_FILE"
      sed -i '' "s|base: '.*'|base: '$BASE_URL/$VERSION/'|g" "$CONFIG_FILE"
  else
      sed -i "s|base: \".*\"|base: \"$BASE_URL/$VERSION/\"|g" "$CONFIG_FILE"
      sed -i "s|base: '.*'|base: '$BASE_URL/$VERSION/'|g" "$CONFIG_FILE"
  fi

  # Build
  pnpm run docs:build
  
  # Move the built assets to the main dist directory under a subfolder
  # The build output is usually .vitepress/dist
  mkdir -p "$DIST_DIR/$VERSION"
  cp -r .vitepress/dist/* "$DIST_DIR/$VERSION/"
  
  echo "Finished building $VERSION"
done

# Cleanup
rm -rf "$TEMP_DIR"

echo "------------------------------------------------"
echo "All versions built successfully!"
echo "Output directory: $DIST_DIR"
