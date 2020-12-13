package loader

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"prod_catalog/services/data"

	"github.com/pkg/errors"
)

type urlLoader struct {
	timeout time.Duration
}

func New(timeout time.Duration) URLLoader {
	return urlLoader{timeout: timeout}
}

func (u urlLoader) Load(ctx context.Context, url string) (chan data.Product, chan error, error) {
	cln := &http.Client{
		Timeout: u.timeout,
	}

	resp, err := cln.Get(url)
	if err != nil {
		return nil, nil, err
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "can not read response body")
		}
		return nil, nil, fmt.Errorf("status: %d, message: %s", resp.StatusCode, string(buf))
	}

	dataChan := make(chan data.Product)
	errChan := make(chan error)

	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Output(0, fmt.Sprintf("%v", err))
			}
		}()
		defer resp.Body.Close()
		defer close(dataChan)
		defer close(errChan)

		r := csv.NewReader(resp.Body)
		r.ReuseRecord = true
		r.Comma = ';'
		r.FieldsPerRecord = 2
		r.TrimLeadingSpace = true

	loop:
		for {
			strs, err := r.Read()
			select {
			case <-ctx.Done():
				break loop
			default:
				switch {
				case err == nil:
					price, err2 := strconv.ParseFloat(strs[1], 64)
					if err2 != nil {
						errChan <- err2
						return
					}
					dataChan <- data.Product{
						Name:  strs[0],
						Price: price,
					}
				case err == io.EOF:
					break loop
				case err != nil:
					errChan <- err
					return
				}
			}
		}

	}()

	return dataChan, errChan, nil
}
