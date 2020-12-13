package api

import (
	"context"
	"sync"

	"prod_catalog/config"
	"prod_catalog/services/data"
	"prod_catalog/services/loader"
	"prod_catalog/services/proto/build/products"
	"prod_catalog/services/storage"

	"github.com/spf13/viper"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/golang/protobuf/ptypes/empty"
)

type service struct {
	products.UnimplementedProductsServer

	storage storage.Storage
	loader  loader.URLLoader
}

type Option func(s *service)

func NewApi(options ...Option) products.ProductsServer {
	s := &service{}
	for _, option := range options {
		option(s)
	}

	return s
}

func WithStorage(st storage.Storage) Option {
	return func(s *service) {
		s.storage = st
	}
}

func WithLoader(l loader.URLLoader) Option {
	return func(s *service) {
		s.loader = l
	}
}

func (s *service) Fetch(ctx context.Context, request *products.FetchRequest) (_ *empty.Empty, resErr error) {
	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	dataChan, errChan, err := s.loader.Load(cancelCtx, request.GetUrl())
	if err != nil {
		return nil, err
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		err := <-errChan
		if err != nil {
			resErr = err
			cancel()
		}
	}()

	go func() {
		defer wg.Done()
		size := viper.GetInt(config.AppBatchSize)
		prodBatch := make([]data.Product, 0, size)
	loop:
		for {
			select {
			case <-cancelCtx.Done():
				break loop
			case prod, ok := <-dataChan:
				if !ok {
					break loop
				}

				prodBatch = append(prodBatch, prod)

				if len(prodBatch) == size {
					if err := s.storage.Store(cancelCtx, prodBatch...); err != nil {
						resErr = err
						cancel()
						return
					}
					prodBatch = prodBatch[:0]
				}
			}
		}

		if len(prodBatch) != 0 {
			if err := s.storage.Store(cancelCtx, prodBatch...); err != nil {
				resErr = err
				cancel()
				return
			}
		}
	}()

	wg.Wait()
	return &empty.Empty{}, resErr
}

func (s *service) List(ctx context.Context, request *products.ListRequest) (*products.ListResponse, error) {
	var sorts []data.OrderParam
	if len(request.GetSort()) != 0 {
		sorts = make([]data.OrderParam, len(request.GetSort()))
		for i, s := range request.GetSort() {
			if s == nil {
				continue
			}
			sorts[i] = data.OrderParam{
				FieldName: s.Field,
				Direction: data.SortingDirection(s.Order),
			}
		}
	}

	result, err := s.storage.List(ctx, sorts, request.GetOffset(), request.GetLimit())
	if err != nil {
		return nil, err
	}

	resp := make([]*products.Product, len(result))
	for i, product := range result {
		resp[i] = &products.Product{
			Name:         product.Name,
			Price:        product.Price,
			UpdatesCount: int64(product.UpdatesCount),
			UpdatedAt:    timestamppb.New(product.UpdatedAt),
		}
	}

	return &products.ListResponse{Items: resp}, nil
}
