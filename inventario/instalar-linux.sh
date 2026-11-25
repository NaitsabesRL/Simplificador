#!/usr/bin/env bash
# ============================================================
#  INSTALADOR — Gestión de Inventario (pollos, carnes, embutidos)
#  Uso:  chmod +x instalar-linux.sh && ./instalar-linux.sh
#  Hace exactamente 3 cosas: copia el programa, crea el comando
#  "inventario" y un acceso directo en el menú de aplicaciones.
# ============================================================
set -e

NOMBRE="Inventario"
BIN_SRC="./inventario-linux"
BIN_NAME="inventario"

echo "==============================================="
echo " Instalación de $NOMBRE para Linux"
echo "==============================================="

# 1) Localizar el binario (junto a este script o en la carpeta actual)
if [ ! -f "$BIN_SRC" ]; then
    if [ -f "./$BIN_NAME" ]; then BIN_SRC="./$BIN_NAME"; else
        echo "ERROR: no se encontró '$BIN_SRC' junto a este script."
        echo "Copia instalar-linux.sh e inventario-linux en la misma carpeta."
        exit 1
    fi
fi

# 2) Elegir destino: /usr/local si hay permisos, si no ~/.local/bin
DESTINO="/usr/local/bin"
if [ ! -w "$DESTINO" ]; then
    if command -v sudo >/dev/null 2>&1; then
        echo "Se usarán permisos de administrador para instalar en $DESTINO."
        sudo cp "$BIN_SRC" "$DESTINO/$BIN_NAME"
        sudo chmod +x "$DESTINO/$BIN_NAME"
    else
        DESTINO="$HOME/.local/bin"
        mkdir -p "$DESTINO"
        cp "$BIN_SRC" "$DESTINO/$BIN_NAME"
        chmod +x "$DESTINO/$BIN_NAME"
    fi
else
    cp "$BIN_SRC" "$DESTINO/$BIN_NAME"
    chmod +x "$DESTINO/$BIN_NAME"
fi
echo " [OK] Programa instalado en: $DESTINO/$BIN_NAME"

# Asegurar que ~/.local/bin esté en el PATH del usuario
case ":$PATH:" in *":$DESTINO:"*) ;;
    *)
        SHE_RC="$HOME/.bashrc"; [ -n "$ZSH_VERSION" ] && SHE_RC="$HOME/.zshrc"
        if ! grep -q "$DESTINO" "$SHELL_RC" 2>/dev/null; then
            echo "export PATH=\"\$PATH:$DESTINO\"" >> "$SHE_RC"
            echo " [OK] Carpeta agregada al PATH (abre una terminal nueva)."
        fi ;;
esac

# 3) Acceso directo en el menú de aplicaciones (entornos gráficos)
DESKTOP="$HOME/.local/share/applications/inventario.desktop"
mkdir -p "$(dirname "$DESKTOP")"
cat > "$DESKTOP" <<EOF
[Desktop Entry]
Type=Application
Name=$NOMBRE
Comment=Gestión de inventario: pollos, carnes y embutidos
Exec=bash -c '$DESTINO/$BIN_NAME; echo; read -p \"Presiona ENTER para cerrar...\"'
Terminal=true
Icon=accessories-calculator
Categories=Office;Finance;
EOF
chmod +x "$DESKTOP"
echo " [OK] Acceso directo creado en el menú de aplicaciones."

# 4) Verificación
"$DESTINO/$BIN_NAME" --version >/dev/null && echo " [OK] Verificado: el programa funciona."

echo ""
echo " ✅ INSTALACIÓN COMPLETA"
echo " ------------------------"
echo " Para abrirlo escribe en la terminal:  inventario"
echo " O búscalo en el menú de aplicaciones como \"$NOMBRE\"."
echo " Tus datos se guardan en: ~/.config/inventario"
echo " Para desinstalar:  rm \"$DESTINO/$BIN_NAME\" \"$DESKTOP\""
