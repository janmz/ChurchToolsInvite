#!/usr/bin/env bash
# Embeds vaya.ico (project root, or assets/vaya.ico) into a Windows .syso resource for go build.
set -euo pipefail

ICON=""
if [ -f vaya.ico ]; then
  ICON="vaya.ico"
elif [ -f assets/vaya.ico ]; then
  ICON="assets/vaya.ico"
else
  echo "No vaya.ico found; skipping Windows icon embed."
  exit 0
fi

VERSION="${1:-0.0.0}"
export GOOS="${GOOS:-windows}"
export GOARCH="${GOARCH:-amd64}"

# Use go run (host-native tool build). go install is fragile in CI: GOBIN/PATH and
# cross-GOOS can leave go-winres.exe instead of a runnable linux binary.
go run github.com/tc-hib/go-winres@v0.3.3 simply \
  --icon "$ICON" \
  --manifest cli \
  --file-version "$VERSION" \
  --product-version "$VERSION"

echo "Embedded $ICON for ${GOOS}/${GOARCH}"
