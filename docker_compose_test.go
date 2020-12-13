package prod_catalog

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"prod_catalog/services/proto/build/products"

	"github.com/stretchr/testify/require"

	"google.golang.org/grpc"
)

func BenchmarkCatalog(b *testing.B) {
	dialCtx, dialCancel := context.WithTimeout(context.Background(), time.Second)
	defer dialCancel()

	conn, err := grpc.DialContext(dialCtx, "localhost:80", grpc.WithInsecure())
	require.NoError(b, err)

	cln := products.NewProductsClient(conn)
	b.ResetTimer()
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < b.N; i++ {
		fCtx, fCancel := context.WithTimeout(context.Background(), time.Second)
		func() {
			idx := rand.Intn(10) + 1
			defer fCancel()
			_, err := cln.Fetch(fCtx, &products.FetchRequest{
				Url: fmt.Sprintf("http://csvsource:3000/csv?idx=%d", idx),
			})
			require.NoError(b, err)
		}()

	}
}
