package storage

import (
	"context"

	"prod_catalog/services/data"
)

type Storage interface {
	List(ctx context.Context, order []data.OrderParam, offset, limit int64) ([]data.Product, error)
	Store(ctx context.Context, products ...data.Product) error
}
