#!/bin/bash

set -e

# Read version from .release-please-manifest.json
VERSION=$(jq -r '.["."]' .release-please-manifest.json)

if [ -z "$VERSION" ] || [ "$VERSION" = "null" ]; then
    echo "❌ Could not extract version from .release-please-manifest.json"
    exit 1
fi

echo "🚀 Syncing version $VERSION..."

if command -v gsed >/dev/null 2>&1; then
  SED_CMD="gsed -i -E"
elif [[ "$OSTYPE" == "darwin"* ]]; then
  SED_CMD="sed -i '' -E"
else
  SED_CMD="sed -i -E"
fi

CHART_DIR="charts/cloud-sentinel-k8s"

# 1. Update README.md (root) - Example: replacing image tags or links
# Assuming we want to replace existing version occurrences.
# We will use a regex to find 0.0.0 or older versions in specific contexts if possible,
# or just generic replacement if that's what the old script did.
# The user asked to update "README.md file", "app version in Charts README.md", "image tag in chart values file", "app version in charts charts.yaml".

# Helper function to update version
update_file() {
    local file=$1
    local search_pattern=$2
    local replace_pattern=$3
    
    if [ -f "$file" ]; then
        echo "Updating $file..."
        $SED_CMD "s|$search_pattern|$replace_pattern|g" "$file"
    else
        echo "⚠️  $file not found"
    fi
}

# Update Root README
# Replace docker tag: ghcr.io/pixelvide/cloud-sentinel-k8s:0.0.0
$SED_CMD "s|(ghcr.io/pixelvide/cloud-sentinel-k8s:)[0-9]+\.[0-9]+\.[0-9]+|\1$VERSION|g" README.md
# Replace helm install URL version: refs/tags/v0.0.0
$SED_CMD "s|(refs/tags/v)[0-9]+\.[0-9]+\.[0-9]+|\1$VERSION|g" README.md

# Update Chart README
# Update app version mentions if any? Or usually just keep in sync.
# Previous script did a simple replace. Let's try to be specific for App Version or Image Tag.
$SED_CMD "s|(ghcr.io/pixelvide/cloud-sentinel-k8s:)[0-9]+\.[0-9]+\.[0-9]+|\1$VERSION|g" "$CHART_DIR/README.md"
# Update Version badge: ![Version: v0.0.0] -> ![Version: v1.0.0]
$SED_CMD "s|(Version: v)[0-9]+\.[0-9]+\.[0-9]+|\1$VERSION|g" "$CHART_DIR/README.md"
# Update Version link: Version-v0.0.0 -> Version-v1.0.0
$SED_CMD "s|(Version-v)[0-9]+\.[0-9]+\.[0-9]+|\1$VERSION|g" "$CHART_DIR/README.md"
# Update AppVersion badge/link
$SED_CMD "s|(AppVersion: v?)[0-9]+\.[0-9]+\.[0-9]+|\1$VERSION|g" "$CHART_DIR/README.md"
$SED_CMD "s|(AppVersion-v?)[0-9]+\.[0-9]+\.[0-9]+|\1$VERSION|g" "$CHART_DIR/README.md"

# Update Chart Values
# tag: 0.0.0
$SED_CMD "s|(tag: )\"?[0-9]+\.[0-9]+\.[0-9]+\"?|\1$VERSION|g" "$CHART_DIR/values.yaml"

# Update Chart.yaml
# version: 0.0.0 -> version: 1.2.3 (Strict SemVer, no v)
$SED_CMD "s|(version: )\"?[0-9]+\.[0-9]+\.[0-9]+\"?|\1$VERSION|g" "$CHART_DIR/Chart.yaml"

# appVersion: "v0.0.0" -> appVersion: "v1.2.3"
# Note: User wanted to strip v in some places, but Chart.yaml appVersion usually keeps v if that's the convention.
# User config had `include-v-in-tag: true`. The previous `release-please` setup with `simple` updater handles this.
# If this script is run INSTEAD or ALONGSIDE, we should match that.
# Matches: appVersion: "v0.0.0" or appVersion: v0.0.0
$SED_CMD "s|(appVersion: \"?v?)[0-9]+\.[0-9]+\.[0-9]+\"?|\1$VERSION\"|g" "$CHART_DIR/Chart.yaml"

echo "✅ Version sync complete!"