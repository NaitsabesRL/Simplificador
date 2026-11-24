// Package storage implementa la persistencia local (JSON) multiplataforma.
// Funciona igual en Linux y Windows: los datos se guardan en el directorio
// de configuración del usuario ($HOME/.inventario o %APPDATA%\inventario).
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ejemplo/inventario/internal/domain"
)

// Datos es la estructura completa del archivo de inventario.
type Datos struct {
	NextProductoID   int64              `json:"next_producto_id"`
	NextMovimientoID int64             `json:"next_movimiento_id"`
	Productos        []domain.Producto  `json:"productos"`
	Movimientos      []domain.Movimiento `json:"movimientos"`
}

// Almacen maneja la lectura/escritura segura del inventario.
type Almacen struct {
	mu     sync.Mutex
	ruta   string
	datos  Datos
}

// RutaDatos devuelve el archivo de datos según el sistema operativo.
// Se puede sobrescribir con la variable de entorno INVENTARIO_HOME.
func RutaDatos() (string, error) {
	if h := os.Getenv("INVENTARIO_HOME"); h != "" {
		if err := os.MkdirAll(h, 0o755); err != nil {
			return "", err
		}
		return filepath.Join(h, "inventario.json"), nil
	}
	dir, err := os.UserConfigDir() // Linux: ~/.config, Windows: %APPDATA%
	if err != nil {
		home, herr := os.UserHomeDir()
		if herr != nil {
			return "", fmt.Errorf("no se pudo determinar el directorio de datos: %w", err)
		}
		dir = home
	}
	base := filepath.Join(dir, "inventario")
	if err := os.MkdirAll(base, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(base, "inventario.json"), nil
}

// Abrir carga (o crea) la base de datos local.
func Abrir(ruta string) (*Almacen, error) {
	a := &Almacen{ruta: ruta}
	b, err := os.ReadFile(ruta)
	if os.IsNotExist(err) {
		a.datos.NextProductoID = 1
		a.datos.NextMovimientoID = 1
		return a, a.guardar()
	}
	if err != nil {
		return nil, err
	}
	if len(b) > 0 {
		if err := json.Unmarshal(b, &a.datos); err != nil {
			// Respalda el archivo corrupto para no perder datos.
			_ = os.Rename(ruta, ruta+".corrupto."+time.Now().Format("20060102-150405"))
			return nil, fmt.Errorf("archivo de datos dañado (se respaldó): %w", err)
		}
	}
	if a.datos.NextProductoID == 0 {
		a.datos.NextProductoID = 1
	}
	if a.datos.NextMovimientoID == 0 {
		a.datos.NextMovimientoID = 1
	}
	return a, nil
}

// guardar escribe de forma atómica (archivo temporal + rename).
func (a *Almacen) guardar() error {
	tmp := a.ruta + ".tmp"
	b, err := json.MarshalIndent(a.datos, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, a.ruta)
}

// Guardar persiste los cambios (thread-safe).
func (a *Almacen) Guardar() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.guardar()
}

// Ruta informa dónde está el archivo de datos (útil para el usuario).
func (a *Almacen) Ruta() string { return a.ruta }

// --- Productos -------------------------------------------------------------

// AgregarProducto valida, asigna ID y persiste un producto nuevo.
func (a *Almacen) AgregarProducto(p *domain.Producto) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := p.Validar(); err != nil {
		return err
	}
	ahora := time.Now()
	p.ID = a.datos.NextProductoID
	a.datos.NextProductoID++
	p.CreadoEn = ahora
	p.ActualizadoEn = ahora
	a.datos.Productos = append(a.datos.Productos, *p)
	return a.guardar()
}

// ListarProductos devuelve todos los productos ordenados por nombre.
func (a *Almacen) ListarProductos() []domain.Producto {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]domain.Producto, len(a.datos.Productos))
	copy(out, a.datos.Productos)
	sort.Slice(out, func(i, j int) bool {
		return strings.ToLower(out[i].Nombre) < strings.ToLower(out[j].Nombre)
	})
	return out
}

// BuscarPorID localiza un producto.
func (a *Almacen) BuscarPorID(id int64) (*domain.Producto, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for i := range a.datos.Productos {
		if a.datos.Productos[i].ID == id {
			p := a.datos.Productos[i]
			return &p, nil
		}
	}
	return nil, domain.ErrNoExiste
}

// BuscarPorNombre busca coincidencias parciales (sin distinguir mayúsculas).
func (a *Almacen) BuscarPorNombre(texto string) []domain.Producto {
	var res []domain.Producto
	t := strings.ToLower(strings.TrimSpace(texto))
	for _, p := range a.ListarProductos() {
		if t == "" || strings.Contains(strings.ToLower(p.Nombre), t) ||
			strings.Contains(strings.ToLower(string(p.Categoria)), t) ||
			strings.Contains(strings.ToLower(p.Proveedor), t) {
			res = append(res, p)
		}
	}
	return res
}

// ActualizarProducto guarda cambios de un producto existente.
func (a *Almacen) ActualizarProducto(p *domain.Producto) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := p.Validar(); err != nil {
		return err
	}
	for i := range a.datos.Productos {
		if a.datos.Productos[i].ID == p.ID {
			p.CreadoEn = a.datos.Productos[i].CreadoEn
			p.ActualizadoEn = time.Now()
			a.datos.Productos[i] = *p
			return a.guardar()
		}
	}
	return domain.ErrNoExiste
}

// EliminarProducto quita un producto y sus movimientos.
func (a *Almacen) EliminarProducto(id int64) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	idx := -1
	for i := range a.datos.Productos {
		if a.datos.Productos[i].ID == id {
			idx = i
			break
		}
	}
	if idx < 0 {
		return domain.ErrNoExiste
	}
	a.datos.Productos = append(a.datos.Productos[:idx], a.datos.Productos[idx+1:]...)
	movs := a.datos.Movimientos[:0]
	for _, m := range a.datos.Movimientos {
		if m.ProductoID != id {
			movs = append(movs, m)
		}
	}
	a.datos.Movimientos = movs
	return a.guardar()
}

// --- Movimientos (entradas / salidas / merma) -------------------------------

// RegistrarMovimiento ajusta el stock y deja trazabilidad del cambio.
func (a *Almacen) RegistrarMovimiento(productoID int64, tipo string, cantidad float64, nota string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	valido := false
	for _, t := range domain.MovimientoTipos {
		if t == tipo {
			valido = true
			break
		}
	}
	if !valido {
		return fmt.Errorf("tipo de movimiento inválido (use: entrada, salida, ajuste, merma)")
	}
	if cantidad <= 0 && tipo != "ajuste" {
		return fmt.Errorf("la cantidad debe ser mayor que cero")
	}
	var prod *domain.Producto
	for i := range a.datos.Productos {
		if a.datos.Productos[i].ID == productoID {
			prod = &a.datos.Productos[i]
			break
		}
	}
	if prod == nil {
		return domain.ErrNoExiste
	}
	switch tipo {
	case "entrada":
		prod.Stock += cantidad
	case "salida", "merma":
		if cantidad > prod.Stock {
			return fmt.Errorf("stock insuficiente: solo hay %.3f %s", prod.Stock, prod.Unidad)
		}
		prod.Stock -= cantidad
	case "ajuste":
		prod.Stock = cantidad
	}
	prod.ActualizadoEn = time.Now()
	a.datos.Movimientos = append(a.datos.Movimientos, domain.Movimiento{
		ID:         a.datos.NextMovimientoID,
		ProductoID: productoID,
		Tipo:       tipo,
		Cantidad:   cantidad,
		Nota:       nota,
		Fecha:      time.Now(),
	})
	a.datos.NextMovimientoID++
	return a.guardar()
}

// HistorialMovimientos devuelve los últimos movimientos de un producto.
func (a *Almacen) HistorialMovimientos(productoID int64, limite int) []domain.Movimiento {
	a.mu.Lock()
	defer a.mu.Unlock()
	var res []domain.Movimiento
	for i := len(a.datos.Movimientos) - 1; i >= 0 && len(res) < limite; i-- {
		if a.datos.Movimientos[i].ProductoID == productoID {
			res = append(res, a.datos.Movimientos[i])
		}
	}
	return res
}

// --- Reportes ---------------------------------------------------------------

// Resumen agrupa estadísticas para la pantalla principal.
type Resumen struct {
	TotalProductos   int
	ValorCosto       float64
	ValorVenta       float64
	BajoStock        []domain.Producto
	PorVencer        []domain.Producto // próximos 3 días o vencidos
	EntradasHoy      int
	SalidasHoy       int
}

// CalcularResumen genera las alertas y totales actuales.
func (a *Almacen) CalcularResumen() Resumen {
	now := time.Now()
	hoy := now.Format("2006-01-02")
	r := Resumen{}
	for _, p := range a.ListarProductos() {
		r.TotalProductos++
		r.ValorCosto += p.Stock * p.PrecioCosto
		r.ValorVenta += p.Stock * p.PrecioVenta
		if p.BajoStock() {
			r.BajoStock = append(r.BajoStock, p)
		}
		if dias, ok := p.DiasParaVencer(now); ok && dias <= 3 {
			r.PorVencer = append(r.PorVencer, p)
		}
	}
	a.mu.Lock()
	for _, m := range a.datos.Movimientos {
		if m.Fecha.Format("2006-01-02") == hoy {
			switch m.Tipo {
			case "entrada":
				r.EntradasHoy++
			case "salida", "merma":
				r.SalidasHoy++
			}
		}
	}
	a.mu.Unlock()
	return r
}

// ExportarCSV genera un reporte CSV compatible con Excel (con BOM UTF-8).
func (a *Almacen) ExportarCSV(destino string) error {
	f, err := os.Create(destino)
	if err != nil {
		return err
	}
	defer f.Close()
	f.WriteString("\xEF\xBB\xBF") // BOM para que Excel abra bien los acentos
	cab := []string{"ID", "Nombre", "Categoria", "Unidad", "Stock", "StockMinimo",
		"PrecioCosto", "PrecioVenta", "Proveedor", "Vencimiento", "Estado"}
	fmt.Fprintln(f, strings.Join(cab, ","))
	for _, p := range a.ListarProductos() {
		venc := ""
		if !p.FechaVencimiento.IsZero() {
			venc = p.FechaVencimiento.Format("2006-01-02")
		}
		estado := "OK"
		if p.BajoStock() {
			estado = "BAJO STOCK"
		}
		if dias, ok := p.DiasParaVencer(time.Now()); ok && dias <= 3 {
			estado += " | POR VENCER"
		}
		fmt.Fprintf(f, "%d;%q;%s;%s;%.3f;%.3f;%.2f;%.2f;%q;%s;%s\n",
			p.ID, p.Nombre, p.Categoria, p.Unidad, p.Stock, p.StockMinimo,
			p.PrecioCosto, p.PrecioVenta, p.Proveedor, venc, estado)
	}
	return nil
}
