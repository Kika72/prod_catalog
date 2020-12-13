package testutils

import (
	"context"
	"testing"

	"prod_catalog/config"
	"prod_catalog/connection"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/stretchr/testify/require"

	"github.com/spf13/viper"
)

const (
	TestMongoCollection = "test_prod_coll"
)

func PrepareMongoDB(t *testing.T, ctx context.Context) *mongo.Database {
	viper.SetDefault(config.MongoHost, "localhost")
	viper.SetDefault(config.MongoPort, 27017)
	viper.SetDefault(config.MongoDB, "product_test")
	viper.SetDefault(config.MongoUsername, "root")
	viper.SetDefault(config.MongoPassword, "RootToor396a")
	viper.SetDefault(config.MongoCollection, TestMongoCollection)

	db, err := connection.GetMongoDatabase(ctx)
	require.NoError(t, err)

	err = db.Collection(TestMongoCollection).Drop(ctx)
	require.NoError(t, err)

	return db
}
