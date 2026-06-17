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
GOOS="${GOOS:-windows}"
GOARCH="${GOARCH:-amd64}"

export GOOS GOARCH
go install github.com/tc-hib/go-winres@v0.3.3

GO_WINRES="$(go env GOPATH)/bin/go-winres"
if [ ! -x "$GO_WINRES" ]; then
  echo "go-winres not found at $GO_WINRES after go install" >&2
  exit 1
fi

"$GO_WINRES" simply \
  --icon "$ICON" \
  --manifest cli \
  --file-version "$VERSION" \
  --product-version "$VERSION"

echo "Embedded $ICON for ${GOOS}/${GOARCH}"
