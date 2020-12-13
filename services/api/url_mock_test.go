package api

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"prod_catalog/services/data"
)

type URLLoaderMock struct {
}

func (U URLLoaderMock) Load(ctx context.Context, s string) (chan data.Product, chan error, error) {
	idx, err := strconv.Atoi(s)
	if err != nil {
		return nil, nil, err
	}
	idx++

	dataChan := make(chan data.Product)
	errChan := make(chan error)

	go func() {
		defer close(dataChan)
		defer close(errChan)

		for i := 0; i < 10; i++ {
			dataChan <- data.Product{
				Name: fmt.Sprintf("name %d;%d\n",
					i*10*idx,
					idx*i,
				),
				Price:        float64(idx * i),
				UpdatedAt:    time.Now().UTC().Add(-time.Hour),
				UpdatesCount: idx,
			}
		}
	}()

	return dataChan, errChan, nil
}
