package api

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"prod_catalog/config"
	"prod_catalog/services/data"
	"prod_catalog/services/proto/build/products"
	"prod_catalog/services/storage"
	"prod_catalog/testutils"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func TestService_List(t *testing.T) {
	ctx := context.Background()
	grpcConn, st := prepareServerAndClient(t)
	defer grpcConn.Close()

	// prepare data
	{
		pp := make([]data.Product, 100)
		for i, _ := range pp {
			pp[i] = data.Product{
				Name:  fmt.Sprintf("name %03d", i+1),
				Price: float64(i*10 + 1),
			}
		}
		require.NoError(t, st.Store(ctx, pp...))
	}

	cln := products.NewProductsClient(grpcConn)
	// first page
	{
		resp, err := cln.List(ctx, &products.ListRequest{
			Offset: 0,
			Limit:  2,
			Sort: []*products.SortParam{
				{
					Field: "name",
					Order: -1,
				},
			},
		})
		require.NoError(t, err)
		require.Len(t, resp.Items, 2)
		require.Equal(t, "name 100", resp.Items[0].Name)
		require.Equal(t, "name 099", resp.Items[1].Name)
	}

	// second page
	{
		resp, err := cln.List(ctx, &products.ListRequest{
			Offset: 2,
			Limit:  2,
			Sort: []*products.SortParam{
				{
					Field: "name",
					Order: -1,
				},
			},
		})
		require.NoError(t, err)
		require.Len(t, resp.Items, 2)
		require.Equal(t, "name 098", resp.Items[0].Name)
		require.Equal(t, "name 097", resp.Items[1].Name)
	}
}

func TestService_Fetch(t *testing.T) {
	ctx := context.Background()
	grpcConn, _ := prepareServerAndClient(t)
	defer grpcConn.Close()

	cln := products.NewProductsClient(grpcConn)
	// "http://localhost:3000/csv?idx=2"
	_, err := cln.Fetch(ctx, &products.FetchRequest{Url: "2"})
	require.NoError(t, err)

	resp, err := cln.List(ctx, &products.ListRequest{
		Offset: 0,
		Limit:  100,
		Sort:   nil,
	})
	require.NoError(t, err)
	require.Len(t, resp.Items, 10)
}

func prepareServerAndClient(t *testing.T) (*grpc.ClientConn, storage.Storage) {
	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cln, mongodb := testutils.PrepareMongoDB(t, context.Background())
	st, err := storage.New(ctx, cln, mongodb.Collection(viper.GetString(config.MongoCollection)))
	require.NoError(t, err)
	api := NewApi(
		WithLoader(&URLLoaderMock{}),
		WithStorage(st),
	)

	products.RegisterProductsServer(s, api)
	go func() {
		if err := s.Serve(lis); err != nil {
			require.NoError(t, err)
		}
	}()

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	grpcConn, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithContextDialer(dialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}

	return grpcConn, st
}
