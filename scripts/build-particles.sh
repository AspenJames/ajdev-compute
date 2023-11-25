#!/usr/bin/env bash

content_dir="$(dirname "$0")/../content/static"
particles_dir="$(dirname "$0")/../internal/particles"

GOOS=js GOARCH=wasm go build -o "$content_dir/particle.wasm" "$particles_dir/main.go"
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" "$content_dir/wasm_exec.js"
