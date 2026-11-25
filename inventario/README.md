# 🐔 Inventario — Pollos, Carnes y Embutidos

Sistema de gestión de inventario escrito en **Go** (sin dependencias externas),
multiplataforma: compila y funciona igual en **Linux** y **Windows**.
Pensado para pollerías, carnicerías y charcuterías: simple, rápido e intuitivo.

## ✨ Qué resuelve (agiliza tu trabajo diario)

| Problema típico | Solución en el programa |
|---|---|
| No saber qué se va a acabar | 🔔 Alertas automáticas de **stock bajo** (comparando con el mínimo que tú defines) |
| Pérdidas por productos caducados | ⏰ Control de **fechas de vencimiento** con aviso de "vence HOY", "vencido" y alerta ≤ 3 días |
| Registros en cuaderno o Excel desordenado | 📥📤 Entradas y salidas con un par de teclas, con **historial de movimientos** (trazabilidad) |
| Mermas (producto dañado) sin control | ⚖️ Registro de **merma y ajustes** con motivo |
| No conocer la ganancia | 📊 Resumen del día: valor del inventario a costo y a venta, **ganancia potencial** |
| Reportes para el contador | 📄 Exportación a **CSV compatible con Excel** (con BOM UTF-8, acentos OK) |
| Buscar un producto tardando | 🔍 Búsqueda rápida por nombre, categoría o proveedor |

## 🚀 Cómo ejecutarlo

### Opción A — Binarios incluidos
- **Linux:** `./inventario-linux`
- **Windows:** doble clic en `inventario.exe` (o ábralo desde una consola)

### Opción B — Compilar tú mismo (necesita Go 1.19+)
```bash
cd inventario
go build -o inventario ./cmd/inventario     # Linux/macOS
go build -o inventario.exe ./cmd/inventario # Windows (o GOOS=windows go build ...)
./inventario
```

### Compilar para ambas plataformas desde una sola máquina
```bash
GOOS=linux   GOARCH=amd64 go build -o inventario-linux ./cmd/inventario
GOOS=windows GOARCH=amd64 go build -o inventario.exe   ./cmd/inventario
```

## 💾 Dónde se guardan los datos

Sin configuración: en el directorio de configuración del usuario.
- Linux: `~/.config/inventario/inventario.json`
- Windows: `%APPDATA%\inventario\inventario.json`
- Personalizable con la variable de entorno `INVENTARIO_HOME` (útil para red/USB).

Las escrituras son **atómicas** (archivo temporal + rename) y si el archivo se
daña, se respalda automáticamente antes de crear uno nuevo: nunca pierdes datos.

## 🧭 Menú principal (todo con números, sin memorizar comandos)

```
 1. 📊 Resumen del día        7. ⚖️  Merma / Ajuste
 2. 📦 Ver inventario         8. 🔔 Alertas (bajo stock y vencimientos)
 3. ➕ Agregar producto       9. 🔍 Buscar producto
 4. ✏️  Editar producto      10. 🗑️  Eliminar producto
 5. 📥 Registrar ENTRADA     11. 📄 Exportar reporte CSV (Excel)
 6. 📤 Registrar SALIDA      12. 🕘 Historial de movimientos
```

Al abrir el programa ya muestra las alertas del día: es lo primero que
necesita saber quien atiende el negocio.

## 🗂️ Estructura del proyecto

```
inventario/
├── cmd/inventario/main.go        # punto de entrada
├── internal/domain/producto.go   # reglas de negocio (producto, categorías, validaciones)
├── internal/storage/almacen.go   # persistencia JSON segura + reportes CSV
└── internal/ui/ui.go             # interfaz de terminal intuitiva
```

## ✅ Validaciones incluidas

- Nombres vacíos, categorías/unidades inválidas → se corrigen en el momento.
- Stock no puede quedar negativo; salida mayor al stock → error claro.
- Precio de venta menor al costo → advertencia de posible pérdida.
- Fechas de vencimiento con formato AAAA-MM-DD verificado.

## 🚀 Instalación fácil (para el usuario final)

Se entregan paquetes listos en `dist/`:

**Windows:** descomprimir `inventario-1.0.0-windows.zip` y **doble clic** en
`instalar-windows.bat`. Sin administrador: copia el programa, crea accesos
directores en el Escritorio y el menú Inicio, y registra el comando `inventario`.

**Linux:** descomprimir `inventario-1.0.0-linux.tar.gz` y ejecutar:
```bash
cd linux && ./instalar-linux.sh
```
Crea el comando `inventario` y un acceso directo en el menú de aplicaciones.

También puede usarse sin instalar: doble clic / `./inventario-linux` directamente.
Flags útiles: `--version`, `--ruta-datos`, `--ayuda`.

Para regenerar los paquetes: `./package.sh`
