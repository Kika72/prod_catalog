package connection

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"prod_catalog/config"

	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetMongoDatabase(ctx context.Context) (*mongo.Client, *mongo.Database, error) {
	host := viper.GetString(config.MongoHost)
	port := viper.GetInt(config.MongoPort)
	address := net.JoinHostPort(host, strconv.Itoa(port))

	clientOptions := options.Client().ApplyURI(fmt.Sprintf("mongodb://%s", address))
	clientOptions.Auth = &options.Credential{
		Username:    viper.GetString(config.MongoUsername),
		Password:    viper.GetString(config.MongoPassword),
		PasswordSet: true,
	}

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, nil, err
	}

	database := client.Database(viper.GetString(config.MongoDB))
	return client, database, nil
}
