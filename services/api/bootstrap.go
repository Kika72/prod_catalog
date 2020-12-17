package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"prod_catalog/config"
	"prod_catalog/connection"
	"prod_catalog/services/loader"
	"prod_catalog/services/proto/build/products"
	"prod_catalog/services/storage"

	"go.mongodb.org/mongo-driver/mongo"

	"google.golang.org/grpc"

	"github.com/spf13/viper"
)

func Run() error {
	// create channel for listening shutdown signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	// create context for signal that shutdown is complete
	ctx, done := context.WithCancel(context.Background())

	l := loader.New(viper.GetDuration(config.AppRequestTimeout))

	connCtx, connCancel := context.WithTimeout(ctx, viper.GetDuration(config.AppRequestTimeout))
	client, mongoDb, err := func() (*mongo.Client, *mongo.Database, error) {
		defer connCancel()
		return connection.GetMongoDatabase(connCtx)
	}()
	if err != nil {
		return err
	}
	st, err := storage.New(connCtx, client, mongoDb.Collection(viper.GetString(config.MongoCollection)))
	if err != nil {
		return err
	}

	api := NewApi(
		WithLoader(l),
		WithStorage(st),
	)
	grpcSrv := grpc.NewServer()
	products.RegisterProductsServer(grpcSrv, api)

	// run shutdown listener
	go func() {
		defer done()
		<-stop

		shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 10*time.Second)
		defer shutdownCancel()
		fmt.Println("shutting down...")
		go func() {
			select {
			case <-shutdownCtx.Done():
				grpcSrv.Stop()
			case <-ctx.Done():
				break
			}
		}()
		grpcSrv.GracefulStop()
		fmt.Println("complete")
	}()

	// create listener
	addr := net.JoinHostPort("0.0.0.0", viper.GetString(config.AppPort))
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	// run grpc service
	go func() {
		fmt.Printf("starting listener at '%s'", addr)
		if err := grpcSrv.Serve(listener); err != nil && err != http.ErrServerClosed {
			fmt.Printf("can not start listener: %v\n", err)
			os.Exit(-1)
		}
	}()

	// waiting shutdown is complete
	<-ctx.Done()
	return nil
}
