#!/bin/bash

# List of target platforms (OS/Architecture)
platforms=(
  "darwin/amd64"
  "darwin/arm64"
  "linux/amd64"
  "linux/arm"
  "linux/arm64"
  "windows/amd64"
)

target="main.go"

version="v1.0.0"

output_dir="build"

mkdir -p $output_dir

for platform in "${platforms[@]}"
do
  IFS="/" read -r GOOS GOARCH <<< "$platform"

  output_name="${output_dir}/app-${version}-${GOOS}-${GOARCH}"
  [[ "$GOOS" == "windows" ]] && output_name+=".exe"

  echo "Building for $GOOS/$GOARCH..."
  
  env GOOS=$GOOS GOARCH=$GOARCH go build -o "$output_name" "$target"

  if [[ $? -ne 0 ]]; then
    echo "Failed to build for $GOOS/$GOARCH!"
    exit 1
  fi

  echo "Successfully built: $output_name"
done

echo "All binaries are built in $output_dir."
