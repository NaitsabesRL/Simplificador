// Package domain contiene las reglas de negocio del sistema de inventario.
package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Categoria de producto (orientado a una pollería / carnicería / charcutería).
type Categoria string

const (
	CatPollos     Categoria = "Pollos"
	CatCarnes     Categoria = "Carnes"
	CatEmbutidos  Categoria = "Embutidos"
	CatLacteos    Categoria = "Lácteos"
	CatCongelados Categoria = "Congelados"
	CatOtros      Categoria = "Otros"
)

// CategoriasValidas lista las categorías disponibles en el menú.
func CategoriasValidas() []Categoria {
	return []Categoria{CatPollos, CatCarnes, CatEmbutidos, CatLacteos, CatCongelados, CatOtros}
}

// EsCategoriaValida verifica si una categoría existe en la lista permitida.
func EsCategoriaValida(c string) bool {
	for _, v := range CategoriasValidas() {
		if strings.EqualFold(string(v), c) {
			return true
		}
	}
	return false
}

// Unidad de medida usada para el stock.
type Unidad string

const (
	Unkg   Unidad = "kg"
	Unlb   Unidad = "lb"
	Unud   Unidad = "unidad"
	Unpack Unidad = "pack"
	Unlitro Unidad = "litro"
)

// UnidadesValidas lista las unidades de medida soportadas.
func UnidadesValidas() []Unidad {
	return []Unidad{Unkg, Unlb, Unud, Unpack, Unlitro}
}

// Producto representa un artículo del inventario.
type Producto struct {
	ID             int64     `json:"id"`
	Nombre         string    `json:"nombre"`
	Categoria      Categoria `json:"categoria"`
	Unidad         Unidad    `json:"unidad"`
	Stock          float64   `json:"stock"`           // cantidad actual
	StockMinimo    float64   `json:"stock_minimo"`    // umbral de alerta
	PrecioCosto    float64   `json:"precio_costo"`    // precio de compra por unidad
	PrecioVenta    float64   `json:"precio_venta"`    // precio de venta por unidad
	Proveedor      string    `json:"proveedor"`
	FechaVencimiento time.Time `json:"fecha_vencimiento,omitempty"` // cero = sin fecha
	CreadoEn       time.Time `json:"creado_en"`
	ActualizadoEn  time.Time `json:"actualizado_en"`
}

// Errores de validación.
var (
	ErrNombreVacio      = errors.New("el nombre no puede estar vacío")
	ErrCategoriaInvalid = errors.New("categoría inválida")
	ErrUnidadInvalid    = errors.New("unidad de medida inválida")
	ErrStockNegativo    = errors.New("el stock no puede ser negativo")
	ErrPrecioNegativo   = errors.New("los precios no pueden ser negativos")
	ErrPrecioInvertido  = errors.New("el precio de venta no puede ser menor al precio de costo")
	ErrNoExiste         = errors.New("el producto no existe")
)

// Validar revisa la coherencia del producto antes de guardarlo.
func (p *Producto) Validar() error {
	if strings.TrimSpace(p.Nombre) == "" {
		return ErrNombreVacio
	}
	if !EsCategoriaValida(string(p.Categoria)) {
		return ErrCategoriaInvalid
	}
	valido := false
	for _, u := range UnidadesValidas() {
		if u == p.Unidad {
			valido = true
			break
		}
	}
	if !valido {
		return ErrUnidadInvalid
	}
	if p.Stock < 0 || p.StockMinimo < 0 {
		return ErrStockNegativo
	}
	if p.PrecioCosto < 0 || p.PrecioVenta < 0 {
		return ErrPrecioNegativo
	}
	if p.PrecioVenta > 0 && p.PrecioCosto > 0 && p.PrecioVenta < p.PrecioCosto {
		return ErrPrecioInvertido
	}
	return nil
}

// BajoStock indica si el producto necesita reposición urgente.
func (p *Producto) BajoStock() bool {
	return p.Stock <= p.StockMinimo
}

// MargenGanancia devuelve el porcentaje de ganancia sobre el costo.
func (p *Producto) MargenGanancia() float64 {
	if p.PrecioCosto == 0 {
		return 0
	}
	return (p.PrecioVenta - p.PrecioCosto) / p.PrecioCosto * 100
}

// DiasParaVencer calcula cuántos días quedan hasta la fecha de vencimiento.
// Devuelve un número negativo si ya venció.
func (p *Producto) DiasParaVencer(now time.Time) (int, bool) {
	if p.FechaVencimiento.IsZero() {
		return 0, false
	}
	d := int(p.FechaVencimiento.Sub(now).Hours() / 24)
	return d, true
}

// EstadoVencimiento describe la situación del producto frente a su caducidad.
func (p *Producto) EstadoVencimiento(now time.Time) string {
	dias, ok := p.DiasParaVencer(now)
	if !ok {
		return "sin fecha"
	}
	switch {
	case dias < 0:
		return fmt.Sprintf("¡VENCIDO hace %d días!", -dias)
	case dias == 0:
		return "vence HOY"
	case dias <= 3:
		return fmt.Sprintf("vence en %d día(s)", dias)
	default:
		return fmt.Sprintf("vence en %d días", dias)
	}
}

// Movimiento registra una entrada o salida de inventario (trazabilidad).
type Movimiento struct {
	ID        int64     `json:"id"`
	ProductoID int64    `json:"producto_id"`
	Tipo      string    `json:"tipo"` // "entrada", "salida", "ajuste", "merma"
	Cantidad  float64   `json:"cantidad"`
	Nota      string    `json:"nota"`
	Fecha     time.Time `json:"fecha"`
}

// MovimientoType tipos válidos de movimiento.
var MovimientoTipos = []string{"entrada", "salida", "ajuste", "merma"}
