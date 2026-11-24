// inventario — Sistema de gestión de inventario para pollos, carnes y
// embutidos. Escrito en Go, funciona en Linux y Windows sin dependencias.
package main

import (
	"fmt"
	"os"

	"github.com/ejemplo/inventario/internal/storage"
	"github.com/ejemplo/inventario/internal/ui"
)

// version se puede fijar al compilar: -ldflags "-X main.version=1.0.0"
var version = "1.0.0"

func main() {
	// Flags sencillos para el instalador / usuario avanzado.
	for _, a := range os.Args[1:] {
		switch a {
		case "--version", "-v", "/version":
			fmt.Println("inventario v" + version)
			return
		case "--ruta-datos":
			p, err := storage.RutaDatos()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}
			fmt.Println(p)
			return
		case "--ayuda", "-h", "/?":
			fmt.Println(`inventario ` + version + ` — Gestión de inventario (pollos, carnes, embutidos)

Uso:
  inventario              Abre el programa (menú interactivo)
  inventario --version    Muestra la versión
  inventario --ruta-datos Muestra dónde se guardan los datos
  inventario --ayuda      Esta ayuda

Variable de entorno opcional:
  INVENTARIO_HOME=<ruta>  Cambia la carpeta donde se guardan los datos`)
			return
		}
	}
	if err := ui.Ejecutar(); err != nil {
		fmt.Fprintln(os.Stderr, "Error fatal:", err)
		os.Exit(1)
	}
}
