package domain

import "errors"

// Producto representa un producto de MercadoLibre
type Producto struct {
	ID          string
	Name        string
	ImageURL    string
	Description string
	Price       float64
	Size        string
	Weight      string
	Color       string
	Rating      float64

	// campos opcionales — puntero distingue entre "no aplica" y "vacío"
	Brand           *string
	ModelVersion    *string
	OS              *string
	BatteryCapacity *string
	Camera          *string
	Memory          *string
	Storage         *string

	// atributos extra sin forma fija
	Specifications map[string]string
}

// NewProducto valida y construye un producto
func NewProducto(name string, price float64) (*Producto, error) {
	if name == "" {
		return nil, ErrNameRequerido
	}
	if price <= 0 {
		return nil, ErrPrecioInvalido
	}
	return &Producto{
		Name:  name,
		Price: price,
	}, nil
}

// StrPtr helper para crear punteros a strings al construir productos
func StrPtr(s string) *string {
	return &s
}

var (
	ErrNameRequerido  = errors.New("name es requerido")
	ErrPrecioInvalido = errors.New("precio debe ser mayor a 0")
	ErrNoEncontrado   = errors.New("producto no encontrado")
)
