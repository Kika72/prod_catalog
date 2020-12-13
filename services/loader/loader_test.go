package loader

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestUrlLoader_Load(t *testing.T) {
	l := New(time.Second)

	dataChan, errChan, err := l.Load(context.Background(), "http://localhost:3000/csv?idx=2")
	require.NoError(t, err)

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		err := <-errChan
		require.NoError(t, err)
	}()

	go func() {
		defer wg.Done()
		count := 0
		for d := range dataChan {
			require.NotEmpty(t, d)
			count++
		}
		require.Equal(t, 10, count)
	}()

	wg.Wait()
}
