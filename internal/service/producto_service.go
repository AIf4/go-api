package service

import (
	"context"
	"fmt"

	"go-meli/internal/domain"
	"go-meli/internal/repository"
)

// ProductoService define qué operaciones existen — el handler depende de esto
type ProductoService interface {
	GetAll(ctx context.Context) ([]*domain.Producto, error)
	GetByID(ctx context.Context, id string) (*domain.Producto, error)
	Create(ctx context.Context, cmd CreateProductoCmd) (*domain.Producto, error)
	Update(ctx context.Context, cmd UpdateProductoCmd) (*domain.Producto, error)
	Delete(ctx context.Context, id string) error
	InsertMany(ctx context.Context, cmds []CreateProductoCmd) ([]*domain.Producto, map[int]error)
}

// CreateProductoCmd contiene los datos necesarios para crear un producto
// — es lo que el handler le pasa al service, sin structs de HTTP ni de Mongo
type CreateProductoCmd struct {
	Name            string
	ImageURL        string
	Description     string
	Price           float64
	Size            string
	Weight          string
	Color           string
	Brand           *string
	ModelVersion    *string
	OS              *string
	BatteryCapacity *string
	Camera          *string
	Memory          *string
	Storage         *string
	Specifications  map[string]string
}

// UpdateProductoCmd contiene los datos para actualizar — incluye el ID
type UpdateProductoCmd struct {
	ID              string
	Name            string
	ImageURL        string
	Description     string
	Price           float64
	Size            string
	Weight          string
	Color           string
	Brand           *string
	ModelVersion    *string
	OS              *string
	BatteryCapacity *string
	Camera          *string
	Memory          *string
	Storage         *string
	Specifications  map[string]string
}

type productoService struct {
	repo repository.ProductoRepository
}

// NewProductoService recibe la interfaz del repositorio, no la implementación
func NewProductoService(repo repository.ProductoRepository) ProductoService {
	return &productoService{repo: repo}
}

func (s *productoService) GetAll(ctx context.Context) ([]*domain.Producto, error) {
	productos, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("obteniendo productos: %w", err)
	}
	return productos, nil
}

func (s *productoService) GetByID(ctx context.Context, id string) (*domain.Producto, error) {
	if id == "" {
		return nil, domain.ErrNoEncontrado
	}

	producto, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return producto, nil
}

/* func (s *productoService) CompareProductos(ctx context.Context, id1, id2 string) (*domain.Producto, *domain.Producto, error) {
	if id1 == "" || id2 == "" {
		return nil, nil, domain.ErrNoEncontrado
	}

} */

func (s *productoService) Create(ctx context.Context, cmd CreateProductoCmd) (*domain.Producto, error) {
	// 1. construye y valida la entidad usando el constructor del dominio
	producto, err := domain.NewProducto(cmd.Name, cmd.Price)
	if err != nil {
		return nil, err // error de dominio: nombre vacío, precio inválido, etc.
	}

	// 2. asigna el resto de campos
	producto.ImageURL = cmd.ImageURL
	producto.Description = cmd.Description
	producto.Size = cmd.Size
	producto.Weight = cmd.Weight
	producto.Color = cmd.Color
	producto.Brand = cmd.Brand
	producto.ModelVersion = cmd.ModelVersion
	producto.OS = cmd.OS
	producto.BatteryCapacity = cmd.BatteryCapacity
	producto.Camera = cmd.Camera
	producto.Memory = cmd.Memory
	producto.Storage = cmd.Storage
	producto.Specifications = cmd.Specifications

	// 3. persiste a través del repositorio
	if err := s.repo.Create(ctx, producto); err != nil {
		return nil, fmt.Errorf("creando producto: %w", err)
	}

	return producto, nil
}

func (s *productoService) InsertMany(ctx context.Context, cmds []CreateProductoCmd) ([]*domain.Producto, map[int]error) {
	if len(cmds) == 0 {
		return nil, nil
	}

	// primero valida todos — si alguno falla, reporta todos los errores
	// sin insertar nada
	prodErrors := make(map[int]error)
	productos := make([]*domain.Producto, 0, len(cmds))

	for i, cmd := range cmds {
		producto, err := domain.NewProducto(cmd.Name, cmd.Price)
		if err != nil {
			prodErrors[i] = err
			continue
		}

		// asigna el resto de campos
		producto.ImageURL = cmd.ImageURL
		producto.Description = cmd.Description
		producto.Size = cmd.Size
		producto.Weight = cmd.Weight
		producto.Color = cmd.Color
		producto.Brand = cmd.Brand
		producto.ModelVersion = cmd.ModelVersion
		producto.OS = cmd.OS
		producto.BatteryCapacity = cmd.BatteryCapacity
		producto.Camera = cmd.Camera
		producto.Memory = cmd.Memory
		producto.Storage = cmd.Storage
		producto.Specifications = cmd.Specifications

		productos = append(productos, producto)
	}

	// si hay errores de validación no inserta nada
	if len(prodErrors) > 0 {
		return nil, prodErrors
	}

	return s.repo.InsertMany(ctx, productos)
}

func (s *productoService) Update(ctx context.Context, cmd UpdateProductoCmd) (*domain.Producto, error) {
	// 1. verifica que el producto existe antes de actualizar
	existing, err := s.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	// 2. aplica los cambios sobre la entidad existente
	existing.Name = cmd.Name
	existing.ImageURL = cmd.ImageURL
	existing.Description = cmd.Description
	existing.Price = cmd.Price
	existing.Size = cmd.Size
	existing.Weight = cmd.Weight
	existing.Color = cmd.Color
	existing.Brand = cmd.Brand
	existing.ModelVersion = cmd.ModelVersion
	existing.OS = cmd.OS
	existing.BatteryCapacity = cmd.BatteryCapacity
	existing.Camera = cmd.Camera
	existing.Memory = cmd.Memory
	existing.Storage = cmd.Storage
	existing.Specifications = cmd.Specifications

	// 3. persiste
	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("actualizando producto: %w", err)
	}

	return existing, nil
}

func (s *productoService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return domain.ErrNoEncontrado
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("eliminando producto: %w", err)
	}

	return nil
}

// En la implementación
