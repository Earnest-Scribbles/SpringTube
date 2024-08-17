package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// Throws an error if the any required environment variables are missing.
	if os.Getenv("PORT") == "" {
		log.Fatal("Please specify the port number for the HTTP server with the environment variable PORT.")
	}

	// Extracts the environment variables.
	PORT := os.Getenv("PORT")

	log.Println("Hello computer!")
	r := gin.Default()

	// ... add route handlers here ...

	// Starts the HTTP server.
	log.Println("Microservice online")
	err := r.Run(":" + PORT)

	if err != nil {
		log.Fatalf("Microservice failed to start: %v", err)
	}

}
