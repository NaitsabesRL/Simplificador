// inventario — Sistema de gestión de inventario para pollos, carnes y
// embutidos. Escrito en Go, funciona en Linux y Windows sin dependencias.
package main

import (
	"fmt"
	"os"

	"github.com/ejemplo/inventario/internal/ui"
)

func main() {
	if err := ui.Ejecutar(); err != nil {
		fmt.Fprintln(os.Stderr, "Error fatal:", err)
		os.Exit(1)
	}
}
