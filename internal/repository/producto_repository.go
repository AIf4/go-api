package repository

import (
	"context"
	"go-meli/internal/domain"
)

type ProductoRepository interface {
	FindAll(ctx context.Context) ([]*domain.Producto, error)
	FindByID(ctx context.Context, id string) (*domain.Producto, error)
	FindInIDs(ctx context.Context, ids []string) (*domain.FindInResult, error)
	Create(ctx context.Context, producto *domain.Producto) error
	Update(ctx context.Context, producto *domain.Producto) error
	Delete(ctx context.Context, id string) error
	InsertMany(ctx context.Context, productos []*domain.Producto) ([]*domain.Producto, map[int]error)
}
