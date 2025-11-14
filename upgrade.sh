#!/usr/bin/env bash
set -e

ROOT_DIR=$(pwd)

go work sync

# Buscar todos los go.mod y recorrerlos
for gomod_file in $(find . -name "go.mod"); do
  dir=$(dirname "$gomod_file")
  echo "🔄 Running go-mod-upgrade in $dir ..."

  cd "$dir"
  go mod download
  go-mod-upgrade
  go mod tidy
  go test ./...
  cd "$ROOT_DIR"
done

echo "✅ Ok."

go work sync