package main

import (
	"log"

	"prod_catalog/config"
	"prod_catalog/services/api"
)

func main() {
	config.Init()
	if err := api.Run(); err != nil {
		log.Fatal(err)
	}
}
