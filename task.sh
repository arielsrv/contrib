#!/usr/bin/env bash
set -euo pipefail

# Directorio raíz desde donde buscar módulos Go
ROOT_DIR=${1:-$(pwd)}

echo "🔍 Buscando módulos Go en $ROOT_DIR..."

# Recorre todas las carpetas que contengan un go.mod
find "$ROOT_DIR" -name "go.mod" | while read -r modfile; do
  dir=$(dirname "$modfile")
  echo "📁 Procesando módulo en: $dir"
  (
    cd "$dir"
    echo "➡️ Ejecutando go mod tidy..."
    go mod tidy

    echo "⬆️ Ejecutando go-mod-upgrade..."
    go-mod-upgrade

    echo "✅ Módulo actualizado: $dir"
    echo "--------------------------------------------"
  )
done

echo "🏁 Todos los módulos procesados correctamente."
