package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

func main() {

	// Confiuration through environment variables is becoming a standard in Microservices
	// Throws an error if the PORT environment variable is missing
	if os.Getenv("PORT") == "" {
		log.Fatal("Please specify the port number for the HTTP server with the environment variable PORT.")
	}

	// Extracts the PORT environment variable.
	PORT := os.Getenv("PORT")

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World!")
	})

	// Registers a HTTP GET route for video streaming.
	r.GET("/video", func(c *gin.Context) {
		videoPath := "./videos/SampleVideo_1280x720_1mb.mp4"

		stats, err := os.Stat(videoPath)
		if err != nil {
			c.String(http.StatusInternalServerError, "File not found")
			return
		}

		c.Header("Content-Length", strconv.Itoa(int(stats.Size())))
		c.Header("Content-Type", "video/mp4")
		c.File(videoPath)
	})

	// Starts the HTTP server.
	log.Printf("Microservice listening on port %s, point your browser at http://localhost:%s/video", PORT, PORT)
	err := r.Run(":" + PORT) // listen and serve on 0.0.0.0:3000

	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
