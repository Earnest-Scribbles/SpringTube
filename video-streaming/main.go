package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {

	// Configuration through environment variables is becoming a standard in Microservices
	// Throws an error if the PORT environment variable is missing
	if os.Getenv("PORT") == "" {
		log.Fatal("Please specify the port number for the HTTP server with the environment variable PORT.")
	}
	if os.Getenv("VIDEO_STORAGE_HOST") == "" {
		log.Fatal("Please specify the host name for the video storage microservice in variable VIDEO_STORAGE_HOST.")
	}
	if os.Getenv("VIDEO_STORAGE_PORT") == "" {
		log.Fatal("Please specify the port number for the video storage microservice in variable VIDEO_STORAGE_PORT.")
	}

	// Extracts the PORT, VIDEO_STORAGE_HOST and VIDEO_STORAGE_PORT environment variable.
	PORT := os.Getenv("PORT")
	VIDEO_STORAGE_HOST := os.Getenv("VIDEO_STORAGE_HOST")
	VIDEO_STORAGE_PORT := os.Getenv("VIDEO_STORAGE_PORT")

	log.Printf("Forwarding video requests to %s:%s.", VIDEO_STORAGE_HOST, VIDEO_STORAGE_PORT)

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World!")
	})

	// Registers a HTTP GET route for video streaming.
	r.GET("/video", func(c *gin.Context) {
		// videoPath := "./videos/SampleVideo_1280x720_1mb.mp4"
		// Created a Director for reverse proxy
		director := func(req *http.Request) {
			req.URL.Host = VIDEO_STORAGE_HOST + ":" + VIDEO_STORAGE_PORT
			req.URL.Path = "/video"
			req.URL.RawQuery = "path=SampleVideo_1280x720_1mb.mp4"
			req.URL.Scheme = "http"
			req.Host = VIDEO_STORAGE_HOST + ":" + VIDEO_STORAGE_PORT
		}

		// Created a reverse proxy
		proxy := &httputil.ReverseProxy{Director: director}
		proxy.ServeHTTP(c.Writer, c.Request)
	})

	// Starts the HTTP server.
	log.Printf("Microservice listening on port %s, point your browser at http://localhost:%s/video", PORT, PORT)
	err := r.Run(":" + PORT) // listen and serve on 0.0.0.0:3000

	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
