#!/usr/bin/env bash

set -euo pipefail

TARGET_OS="${GOOS:-darwin}"
TARGET_ARCH="${GOARCH:-$(go env GOARCH)}"
TARGET_CGO="${CGO_ENABLED:-0}"
OUTPUT_DIR="bin"

mkdir -p "${OUTPUT_DIR}"

while IFS= read -r main_file; do
  command_dir="${main_file%/main.go}"
  relative_name="${command_dir#cmd/}"
  binary_name="${relative_name//\//-}"
  output_path="${OUTPUT_DIR}/${binary_name}"

  echo "Building ${command_dir} -> ${output_path} (${TARGET_OS}/${TARGET_ARCH})"
  GOOS="${TARGET_OS}" GOARCH="${TARGET_ARCH}" CGO_ENABLED="${TARGET_CGO}" \
    go build -o "${output_path}" "./${command_dir}"
done < <(find cmd -name main.go -type f | sort)
