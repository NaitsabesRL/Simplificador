#!/usr/bin/env bash
# Empaqueta los instaladores listos para repartir al usuario final.
set -e
cd "$(dirname "$0")"

VER=$(grep -oP 'var version = "\K[^"]+' cmd/inventario/main.go)
OUT="dist"
rm -rf "$OUT"; mkdir -p "$OUT/linux" "$OUT/windows"

# Compilar ambos binarios (optimizados)
go build -ldflags "-s -w -X main.version=$VER" -o inventario-linux ./cmd/inventario
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -X main.version=$VER" -o inventario.exe ./cmd/inventario

# Paquete Linux
cp inventario-linux instalar-linux.sh README.md "$OUT/linux/"
chmod +x "$OUT/linux/instalar-linux.sh"
tar -czf "$OUT/inventario-$VER-linux.tar.gz" -C "$OUT" linux

# Paquete Windows
cp inventario.exe instalar-windows.bat README.md "$OUT/windows/"
(cd "$OUT" && zip -qr "inventario-$VER-windows.zip" windows) || \
  (cd "$OUT/windows" && python3 -c "import zipfile,os; z=zipfile.ZipFile('../inventario-$VER-windows.zip','w',zipfile.ZIP_DEFLATED); [z.write(f) for f in os.listdir('.')]; z.close()")

echo ""
echo "Paquetes creados en $OUT/:"
ls -lh "$OUT"/*.tar.gz "$OUT"/*.zip
