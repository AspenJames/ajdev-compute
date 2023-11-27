#!/usr/bin/env bash
set -e

proj_root="$(realpath "$(dirname "$0")/../")"

# Build WASM particle animation.
GOOS=js GOARCH=wasm go build -o "$proj_root/content/static/particle.wasm" "$proj_root/internal/particles/main.go"
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" "$proj_root/content/static/wasm_exec.js"

# Build CSS.
pushd "$proj_root/content" || exit 1
    npm i && npm run build-css
popd || exit 1

# Build Fastly Compute binary.
GOOS=wasip1 GOARCH=wasm go build -o "$proj_root/bin/main.wasm" "$proj_root/main.go"
