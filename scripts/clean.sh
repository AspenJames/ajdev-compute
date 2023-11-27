#!/usr/bin/env bash
set -xe

proj_root="$(realpath "$(dirname "$0")/../")"

# Clean WASM particle animation.
rm "$proj_root/content/static/particle.wasm"
rm "$proj_root/content/static/wasm_exec.js"

# Clean CSS.
rm "$proj_root/content/static/site.css"
rm -r "$proj_root/content/node_modules"

# Clean Fastly Compute binary & pkg.
rm "$proj_root/bin/main.wasm"
rm "$proj_root/pkg/ajdev-compute.tar.gz"
