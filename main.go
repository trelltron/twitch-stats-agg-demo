package main

import (
	"os"

	"github.com/joho/godotenv"

	"github.com/trelltron/twitch-stats-agg-demo/routes"
	"github.com/trelltron/twitch-stats-agg-demo/services"
)

func main() {
	godotenv.Load()
	services := services.BuildServices()
	router := routes.BuildRouter(&services)

	address, exists := os.LookupEnv("SERVER_ADDRESS")
	if !exists {
		address = "localhost:3000"
	}

	router.Run(address)
}
