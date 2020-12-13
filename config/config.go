package config

import (
	"flag"
	"log"
	"os"

	"github.com/spf13/viper"
)

var (
	flagCfgFile string
)

const (
	AppPort           string = "app.port"
	AppRequestTimeout string = "app.requestTimeout"
	AppBatchSize      string = "app.batchSize"

	MongoHost       string = "mongo.host"
	MongoPort       string = "mongo.port"
	MongoUsername   string = "mongo.username"
	MongoPassword   string = "mongo.password"
	MongoDB         string = "mongo.database"
	MongoCollection string = "mongo.collection"
)

func Init() {
	flag.StringVar(&flagCfgFile, "config", "", "path to config file")
	flag.Parse()

	if flagCfgFile == "" {
		flag.Usage()
		os.Exit(-1)
	}

	viper.SetConfigFile(flagCfgFile)
	if err := viper.ReadInConfig(); err == nil {
		log.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		log.Fatal(err)
	}
}
