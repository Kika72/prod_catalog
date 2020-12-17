package storage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"prod_catalog/config"
	"prod_catalog/services/data"
	"prod_catalog/testutils"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/spf13/viper"

	"github.com/stretchr/testify/require"
)

func TestProductStorage_Store(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cln, db := testutils.PrepareMongoDB(t, ctx)
	coll := db.Collection(viper.GetString(config.MongoCollection))

	storage, err := New(ctx, cln, coll)
	require.NoError(t, err)

	// store new
	{
		p := data.Product{
			Name:  "name 1",
			Price: 123.23,
		}
		require.NoError(t, storage.Store(ctx, p))

		res := db.Collection(viper.GetString(config.MongoCollection)).
			FindOne(ctx, bson.M{"name": bson.M{"$eq": p.Name}})
		require.NoError(t, res.Err())
		saved := data.Product{}
		require.NoError(t, res.Decode(&saved))
		require.Equal(t, p.Name, saved.Name)
		require.Equal(t, p.Price, saved.Price)
		require.WithinDuration(t, time.Now().UTC(), saved.UpdatedAt, 500*time.Millisecond)
		require.Equal(t, 1, saved.UpdatesCount)
	}

	// update existing
	{
		p := data.Product{
			Name:  "name 1",
			Price: 100.1,
		}
		require.NoError(t, storage.Store(ctx, p))

		res, err := db.Collection(viper.GetString(config.MongoCollection)).
			Find(ctx, bson.M{"name": bson.M{"$eq": p.Name}})
		require.NoError(t, err)
		require.NoError(t, res.Err())

		docs := make([]data.Product, 2)
		require.NoError(t, res.All(ctx, &docs))
		require.Len(t, docs, 1)

		saved := docs[0]
		require.Equal(t, p.Name, saved.Name)
		require.Equal(t, p.Price, saved.Price)
		require.WithinDuration(t, time.Now().UTC(), saved.UpdatedAt, 500*time.Millisecond)
		require.Equal(t, 2, saved.UpdatesCount)
	}

	// insert second
	{
		p := data.Product{
			Name:  "name 2",
			Price: 200,
		}
		require.NoError(t, storage.Store(ctx, p))

		res, err := db.Collection(viper.GetString(config.MongoCollection)).
			Find(ctx, bson.D{})
		require.NoError(t, err)

		docs := make([]bson.M, 2)
		require.NoError(t, res.All(ctx, &docs))
		require.Len(t, docs, 2)
	}

	// update second with same price
	{
		p := data.Product{
			Name:  "name 2",
			Price: 200,
		}
		require.NoError(t, storage.Store(ctx, p))

		res, err := db.Collection(viper.GetString(config.MongoCollection)).
			Find(ctx, bson.M{"name": bson.M{"$eq": p.Name}})
		require.NoError(t, err)

		docs := make([]data.Product, 2)
		require.NoError(t, res.All(ctx, &docs))
		require.Len(t, docs, 1)
		saved := docs[0]
		require.Equal(t, 1, saved.UpdatesCount)
	}

	// bulk write
	{
		pp := make([]data.Product, 100)
		for i, _ := range pp {
			pp[i] = data.Product{
				Name:  fmt.Sprintf("name %03d", i+1),
				Price: float64(i*10 + 1),
			}
		}
		require.NoError(t, storage.Store(ctx, pp...))

		res, err := db.Collection(viper.GetString(config.MongoCollection)).
			Find(ctx, bson.D{})
		require.NoError(t, err)

		docs := make([]bson.M, 2)
		require.NoError(t, res.All(ctx, &docs))
		require.Len(t, docs, 102)
	}
}

func TestProductStorage_List(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cln, db := testutils.PrepareMongoDB(t, ctx)
	coll := db.Collection(viper.GetString(config.MongoCollection))

	storage, err := New(ctx, cln, coll)
	require.NoError(t, err)

	// prepare data
	{
		pp := make([]data.Product, 100)
		for i, _ := range pp {
			pp[i] = data.Product{
				Name:  fmt.Sprintf("name %03d", i+1),
				Price: float64(i*10 + 1),
			}
		}
		require.NoError(t, storage.Store(ctx, pp...))
	}

	for i := int64(0); i <= 10; i++ {
		list, err := storage.List(ctx, []data.OrderParam{{FieldName: "name", Direction: 1}}, i*10, 10)
		require.NoError(t, err)
		if i < 10 {
			require.Len(t, list, 10)
			require.Equal(t, fmt.Sprintf("name %03d", i*10+1), list[0].Name)
			require.Equal(t, fmt.Sprintf("name %03d", i*10+int64(len(list))), list[len(list)-1].Name)
		} else {
			require.Len(t, list, 0)
		}
	}

}
