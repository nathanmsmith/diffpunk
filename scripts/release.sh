#!/bin/sh
set -eu
version="${1:-dev}"
mkdir -p release
for task_target_os in darwin linux; do
  for task_target_arch in amd64 arm64; do
    task_release_name="diff-punk_${version}_${task_target_os}_${task_target_arch}"
    task_release_dir="release/$task_release_name"
    mkdir -p "$task_release_dir"
    CGO_ENABLED=0 GOOS="$task_target_os" GOARCH="$task_target_arch" \
      go build -trimpath -ldflags="-s -w -X main.version=$version" -o "$task_release_dir/diff-punk" .
    cp README.md "$task_release_dir/README.md"
    cp internal/viewer/dist/THIRD_PARTY_LICENSES.txt "$task_release_dir/THIRD_PARTY_LICENSES.txt"
    tar -czf "release/$task_release_name.tar.gz" -C "$task_release_dir" .
    rm -rf "$task_release_dir"
  done
done
(cd release && shasum -a 256 diff-punk_"$version"_*.tar.gz > "checksums_${version}.txt")
