// Package ui implementa la interfaz de terminal interactiva y amigable.
package ui

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ejemplo/inventario/internal/domain"
	"github.com/ejemplo/inventario/internal/storage"
)

// App une la interfaz con el almacenamiento.
type App struct {
	almacen *storage.Almacen
	in      *bufio.Reader
}

// Ejecutar inicia el programa (se llama desde main).
func Ejecutar() error {
	ruta, err := storage.RutaDatos()
	if err != nil {
		return err
	}
	a, err := storage.Abrir(ruta)
	if err != nil {
		return err
	}
	app := &App{almacen: a, in: bufio.NewReader(os.Stdin)}
	app.bienvenida()
	for {
		op := app.menu()
		if app.finEntrada(op) {
			fmt.Println("\n(Entrada finalizada. Los datos quedaron guardados.)")
			return nil
		}
		switch op {
		case "1":
			app.pantallaResumen()
		case "2":
			app.pantallaLista()
		case "3":
			app.pantallaAgregar()
		case "4":
			app.pantallaEditar()
		case "5":
			app.pantallaMovimiento("entrada")
		case "6":
			app.pantallaMovimiento("salida")
		case "7":
			app.pantallaMermaAjuste()
		case "8":
			app.pantallaAlertas()
		case "9":
			app.pantallaBuscar()
		case "10":
			app.pantallaEliminar()
		case "11":
			app.pantallaExportar()
		case "12":
			app.pantallaHistorial()
		case "0", "q", "salir":
			fmt.Println("\n¡Hasta luego! Los datos quedaron guardados en:", a.Ruta())
			return nil
		default:
			fmt.Println("Opción no reconocida, intente de nuevo.")
		}
	}
}

// --- utilidades de entrada ---------------------------------------------------

func (app *App) leerLinea(prompt string) string {
	fmt.Print(prompt)
	linea, err := app.in.ReadString('\n')
	if err != nil && linea == "" {
		return "\x04" // señal de fin de entrada (tubería cerrada / Ctrl+D)
	}
	return strings.TrimSpace(linea)
}

// finEntrada indica que ya no hay más entrada del usuario: salir sin bucles.
func (app *App) finEntrada(v string) bool { return v == "\x04" }

func (app *App) leerTexto(prompt, defecto string) string {
	v := app.leerLinea(prompt)
	if v == "" {
		return defecto
	}
	return v
}

func (app *App) leerFloat(prompt string, defecto float64) (float64, bool) {
	v := app.leerLinea(prompt)
	if v == "" {
		return defecto, true
	}
	v = strings.ReplaceAll(v, ",", ".")
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		fmt.Printf("  ⚠ '%s' no es un número válido.\n", v)
		return 0, false
	}
	return f, true
}

func (app *App) pausar() {
	fmt.Print("\nPresione ENTER para continuar...")
	_, _ = app.in.ReadString('\n')
}

func (app *App) bienvenida() {
	fmt.Println(strings.Repeat("=", 58))
	fmt.Println("   🐔  INVENTARIO — Pollos, Carnes y Embutidos  🥩")
	fmt.Println(strings.Repeat("=", 58))
	fmt.Println("  Datos guardados en:", app.almacen.Ruta())
	// Alertas inmediatas al abrir: lo que el usuario necesita saber hoy.
	r := app.almacen.CalcularResumen()
	if len(r.PorVencer) > 0 {
		fmt.Printf("  🔴 ATENCIÓN: %d producto(s) por vencer o vencidos.\n", len(r.PorVencer))
	}
	if len(r.BajoStock) > 0 {
		fmt.Printf("  🟡 Aviso: %d producto(s) con stock bajo.\n", len(r.BajoStock))
	}
}

func (app *App) menu() string {
	fmt.Print(`
┌────────────── MENÚ PRINCIPAL ──────────────┐
│  1. 📊 Resumen del día                     │
│  2. 📦 Ver inventario completo             │
│  3. ➕ Agregar producto nuevo               │
│  4. ✏️  Editar producto                     │
│  5. 📥 Registrar ENTRADA (compra/recepción)│
│  6. 📤 Registrar SALIDA (venta)            │
│  7. ⚖️  Merma / Ajuste de inventario       │
│  8. 🔔 Alertas (bajo stock y vencimientos) │
│  9. 🔍 Buscar producto                     │
│ 10. 🗑️  Eliminar producto                  │
│ 11. 📄 Exportar reporte CSV (Excel)        │
│ 12. 🕘 Historial de movimientos            │
│  0. 🚪 Salir                               │
└────────────────────────────────────────────┘
`)
	return app.leerLinea("Elija una opción: ")
}

// elegirProducto muestra una lista numerada y devuelve el producto elegido.
func (app *App) elegirProducto(titulo string) (*domain.Producto, bool) {
	lista := app.almacen.ListarProductos()
	if len(lista) == 0 {
		fmt.Println("No hay productos registrados. Agregue uno primero (opción 3).")
		return nil, false
	}
	fmt.Println("\n" + titulo)
	app.tabla(lista)
	v := app.leerLinea("Número del producto (ENTER para cancelar): ")
	if v == "" {
		return nil, false
	}
	n, err := strconv.Atoi(v)
	if app.finEntrada(v) || err != nil || n < 1 || n > len(lista) {
		fmt.Println("Número fuera de rango.")
		return nil, false
	}
	p, err := app.almacen.BuscarPorID(lista[n-1].ID)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, false
	}
	return p, true
}

func (app *App) tabla(productos []domain.Producto) {
	if len(productos) == 0 {
		fmt.Println("(sin resultados)")
		return
	}
	now := time.Now()
	fmt.Printf("%-4s %-26s %-12s %10s %-8s %10s %10s %-22s\n",
		"#", "PRODUCTO", "CATEGORÍA", "STOCK", "UND", "COSTO", "VENTA", "VENCIMIENTO")
	fmt.Println(strings.Repeat("-", 108))
	for i, p := range productos {
		marcador := "  "
		if p.BajoStock() {
			marcador = "⚠️"
		}
		fmt.Printf("%-4d %-26s %-12s %10.3f %-8s %10.2f %10.2f %-2s%-20s\n",
			i+1, trunc(p.Nombre, 26), p.Categoria, p.Stock, p.Unidad,
			p.PrecioCosto, p.PrecioVenta, marcador, p.EstadoVencimiento(now))
	}
}

func trunc(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

// --- pantallas ----------------------------------------------------------------

func (app *App) pantallaResumen() {
	r := app.almacen.CalcularResumen()
	fmt.Println("\n===== RESUMEN DEL DÍA =====")
	fmt.Printf("  Productos registrados : %d\n", r.TotalProductos)
	fmt.Printf("  Valor inventario (costo): %.2f\n", r.ValorCosto)
	fmt.Printf("  Valor inventario (venta): %.2f\n", r.ValorVenta)
	fmt.Printf("  Ganancia potencial      : %.2f\n", r.ValorVenta-r.ValorCosto)
	fmt.Printf("  Movimientos hoy         : %d entradas, %d salidas/mermas\n",
		r.EntradasHoy, r.SalidasHoy)
	fmt.Printf("  ⚠️  Bajo stock           : %d producto(s)\n", len(r.BajoStock))
	fmt.Printf("  🔴 Por vencer / vencidos : %d producto(s)\n", len(r.PorVencer))
	if len(r.BajoStock) > 0 || len(r.PorVencer) > 0 {
		fmt.Println("\n  Use la opción 8 para ver las alertas detalladas.")
	}
	app.pausar()
}

func (app *App) pantallaLista() {
	fmt.Println("\n===== INVENTARIO COMPLETO =====")
	app.tabla(app.almacen.ListarProductos())
	app.pausar()
}

func (app *App) pantallaBuscar() {
	q := app.leerLinea("\nTexto a buscar (nombre, categoría o proveedor): ")
	res := app.almacen.BuscarPorNombre(q)
	fmt.Println("\n===== RESULTADOS =====")
	app.tabla(res)
	app.pausar()
}

func (app *App) leerProductoFormulario() domain.Producto {
	var p domain.Producto
	p.Nombre = app.leerTexto("  Nombre del producto: ", "")
	fmt.Println("  Categorías: 1.Pollos 2.Carnes 3.Embutidos 4.Lácteos 5.Congelados 6.Otros")
	c := app.leerLinea("  Elija categoría [1-6]: ")
	cats := domain.CategoriasValidas()
	if n, err := strconv.Atoi(c); err == nil && n >= 1 && n <= len(cats) {
		p.Categoria = cats[n-1]
	} else {
		p.Categoria = domain.CatOtros
	}
	fmt.Println("  Unidades: 1.kg 2.lb 3.unidad 4.pack 5.litro")
	u := app.leerLinea("  Elija unidad [1-5]: ")
	units := domain.UnidadesValidas()
	if n, err := strconv.Atoi(u); err == nil && n >= 1 && n <= len(units) {
		p.Unidad = units[n-1]
	} else {
		p.Unidad = domain.Unkg
	}
	if v, ok := app.leerFloat("  Stock actual [0]: ", 0); ok {
		p.Stock = v
	}
	if v, ok := app.leerFloat("  Stock mínimo para alerta [0]: ", 0); ok {
		p.StockMinimo = v
	}
	if v, ok := app.leerFloat("  Precio de costo [0]: ", 0); ok {
		p.PrecioCosto = v
	}
	if v, ok := app.leerFloat("  Precio de venta [0]: ", 0); ok {
		p.PrecioVenta = v
	}
	p.Proveedor = app.leerTexto("  Proveedor (opcional): ", "")
	if f := app.leerTexto("  Fecha vencimiento AAAA-MM-DD (opcional): ", ""); f != "" {
		if t, err := time.Parse("2006-01-02", f); err == nil {
			p.FechaVencimiento = t
		} else {
			fmt.Println("  ⚠ Fecha no válida, se ignoró (use AAAA-MM-DD).")
		}
	}
	return p
}

func (app *App) pantallaAgregar() {
	fmt.Println("\n===== AGREGAR PRODUCTO =====")
	p := app.leerProductoFormulario()
	if err := app.almacen.AgregarProducto(&p); err != nil {
		fmt.Println("  ❌ No se pudo guardar:", err)
	} else {
		fmt.Printf("  ✅ Producto \"%s\" agregado con ID %d.\n", p.Nombre, p.ID)
	}
	app.pausar()
}

func (app *App) pantallaEditar() {
	p, ok := app.elegirProducto("===== EDITAR PRODUCTO =====")
	if !ok {
		return
	}
	fmt.Printf("\nEditando: %s (deje un campo vacío para conservarlo)\n", p.Nombre)
	p.Nombre = app.leerTexto(fmt.Sprintf("  Nombre [%s]: ", p.Nombre), p.Nombre)
	if v, ok := app.leerFloat(fmt.Sprintf("  Stock actual [%.3f]: ", p.Stock), p.Stock); ok {
		p.Stock = v
	}
	if v, ok := app.leerFloat(fmt.Sprintf("  Stock mínimo [%.3f]: ", p.StockMinimo), p.StockMinimo); ok {
		p.StockMinimo = v
	}
	if v, ok := app.leerFloat(fmt.Sprintf("  Precio costo [%.2f]: ", p.PrecioCosto), p.PrecioCosto); ok {
		p.PrecioCosto = v
	}
	if v, ok := app.leerFloat(fmt.Sprintf("  Precio venta [%.2f]: ", p.PrecioVenta), p.PrecioVenta); ok {
		p.PrecioVenta = v
	}
	p.Proveedor = app.leerTexto(fmt.Sprintf("  Proveedor [%s]: ", p.Proveedor), p.Proveedor)
	if f := app.leerTextoSoloSiCambia("  Fecha vencimiento AAAA-MM-DD (ENTER para mantener): "); f != "" {
		if t, err := time.Parse("2006-01-02", f); err == nil {
			p.FechaVencimiento = t
		} else {
			fmt.Println("  ⚠ Fecha no válida, se mantuvo la anterior.")
		}
	}
	if err := app.almacen.ActualizarProducto(p); err != nil {
		fmt.Println("  ❌ No se pudo actualizar:", err)
	} else {
		fmt.Println("  ✅ Producto actualizado.")
	}
	app.pausar()
}

func (app *App) leertextoSC(prompt string) string { return app.leerLinea(prompt) }

func (app *App) leerTextoSoloSiCambia(prompt string) string {
	return app.leerLinea(prompt)
}

func (app *App) pantallaMovimiento(tipo string) {
	nombre := "ENTRADA"
	if tipo == "salida" {
		nombre = "SALIDA / VENTA"
	}
	p, ok := app.elegirProducto("===== REGISTRAR " + nombre + " =====")
	if !ok {
		return
	}
	fmt.Printf("  %s — stock actual: %.3f %s\n", p.Nombre, p.Stock, p.Unidad)
	for {
		v, ok := app.leerFloat(fmt.Sprintf("  Cantidad en %s: ", p.Unidad), -1)
		if !ok {
			continue
		}
		if v <= 0 {
			fmt.Println("  La cantidad debe ser mayor que cero.")
			continue
		}
		nota := app.leerTexto("  Nota (cliente, proveedor, motivo — opcional): ", "")
		if err := app.almacen.RegistrarMovimiento(p.ID, tipo, v, nota); err != nil {
			fmt.Println("  ❌", err)
			return
		}
		p, _ = app.almacen.BuscarPorID(p.ID)
		fmt.Printf("  ✅ Registrado. Nuevo stock: %.3f %s\n", p.Stock, p.Unidad)
		if p.BajoStock() {
			fmt.Printf("  ⚠️  ALERTA: %s quedó con stock bajo (mínimo %.3f). ¡Hora de reponer!\n",
				p.Nombre, p.StockMinimo)
		}
		if app.leerLinea("  ¿Registrar otra cantidad del mismo producto? (s/N): ") == "s" {
			continue
		}
		break
	}
	app.pausar()
}

func (app *App) pantallaMermaAjuste() {
	fmt.Println("\n===== MERMA / AJUSTE =====")
	fmt.Println("  Merma  = producto dañado, echado a perder o caducado (resta stock)")
	fmt.Println("  Ajuste = corregir el stock manualmente a una cantidad exacta")
	tipo := app.leerLinea("  Elija (merma / ajuste): ")
	if app.finEntrada(tipo) || (tipo != "merma" && tipo != "ajuste") {
		fmt.Println("Opción inválida.")
		app.pausar()
		return
	}
	p, ok := app.elegirProducto("")
	if !ok {
		return
	}
	fmt.Printf("  %s — stock actual: %.3f %s\n", p.Nombre, p.Stock, p.Unidad)
	var v float64
	for {
		x, ok := app.leerFloat("  Cantidad: ", -1)
		if ok && x >= 0 {
			v = x
			break
		}
		fmt.Print("  Cantidad inválida, intente de nuevo.\n")
	}
	nota := app.leerTexto("  Motivo (ej: caducado, error de conteo): ", "")
	if err := app.almacen.RegistrarMovimiento(p.ID, tipo, v, nota); err != nil {
		fmt.Println("  ❌", err)
	} else {
		p, _ = app.almacen.BuscarPorID(p.ID)
		fmt.Printf("  ✅ Hecho. Stock actual: %.3f %s\n", p.Stock, p.Unidad)
	}
	app.pausar()
}

func (app *App) pantallaAlertas() {
	r := app.almacen.CalcularResumen()
	now := time.Now()
	fmt.Println("\n===== 🔔 ALERTAS =====")
	fmt.Println("\n-- Stock bajo (necesita reposición) --")
	if len(r.BajoStock) == 0 {
		fmt.Println("  Todo el inventario tiene stock suficiente. 👍")
	} else {
		for _, p := range r.BajoStock {
			fmt.Printf("  ⚠️  %-26s stock %.3f %s (mínimo %.3f)\n",
				p.Nombre, p.Stock, p.Unidad, p.StockMinimo)
		}
	}
	fmt.Println("\n-- Productos por vencer (≤ 3 días) o vencidos --")
	if len(r.PorVencer) == 0 {
		fmt.Println("  Ningún producto próximo a vencer. 👍")
	} else {
		for _, p := range r.PorVencer {
			fmt.Printf("  🔴 %-26s %-24s stock %.3f %s\n",
				p.Nombre, p.EstadoVencimiento(now), p.Stock, p.Unidad)
		}
		fmt.Println("\n  Sugerencia: haga ofertas o registre merma (opción 7).")
	}
	app.pausar()
}

func (app *App) pantallaEliminar() {
	p, ok := app.elegirProducto("===== ELIMINAR PRODUCTO =====")
	if !ok {
		return
	}
	conf := app.leerLinea(fmt.Sprintf("  ¿Seguro que desea eliminar \"%s\"? escriba SI: ", p.Nombre))
	if conf != "SI" {
		fmt.Println("  Operación cancelada.")
		app.pausar()
		return
	}
	if err := app.almacen.EliminarProducto(p.ID); err != nil {
		fmt.Println("  ❌", err)
	} else {
		fmt.Println("  ✅ Producto eliminado.")
	}
	app.pausar()
}

func (app *App) pantallaExportar() {
	dir, _ := os.Getwd()
	nombre := fmt.Sprintf("inventario_%s.csv", time.Now().Format("2006-01-02"))
	destino := app.leerTexto(fmt.Sprintf("  Guardar en [%s]: ", filepath.Join(dir, nombre)), "")
	if destino == "" {
		destino = filepath.Join(dir, nombre)
	}
	if err := app.almacen.ExportarCSV(destino); err != nil {
		fmt.Println("  ❌ No se pudo exportar:", err)
	} else {
		fmt.Println("  ✅ Reporte exportado a:", destino)
		fmt.Println("     Ábralo con Excel o LibreOffice Calc.")
	}
	app.pausar()
}

func (app *App) pantallaHistorial() {
	p, ok := app.elegirProducto("===== HISTORIAL DE MOVIMIENTOS =====")
	if !ok {
		return
	}
	movs := app.almacen.HistorialMovimientos(p.ID, 20)
	fmt.Printf("\n  Últimos movimientos de %s:\n", p.Nombre)
	if len(movs) == 0 {
		fmt.Println("  (ninguno)")
	}
	for _, m := range movs {
		flecha := "➕"
		if m.Tipo == "salida" || m.Tipo == "merma" {
			flecha = "➖"
		}
		if m.Tipo == "ajuste" {
			flecha = "⚖️"
		}
		nota := ""
		if m.Nota != "" {
			nota = " — " + m.Nota
		}
		fmt.Printf("  %s %s %s %.3f%s (%s)\n", flecha,
			m.Fecha.Format("02/01/2006 15:04"), m.Tipo, m.Cantidad, nota, p.Unidad)
	}
	app.pausar()
}
