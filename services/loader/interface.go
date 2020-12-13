package loader

import (
	"context"

	"prod_catalog/services/data"
)

type URLLoader interface {
	Load(context.Context, string) (chan data.Product, chan error, error)
}
