package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type VideoRecord struct {
	ID        string `bson:"_id"`
	VideoPath string `bson:"videoPath"`
}

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
	if os.Getenv("DBHOST") == "" {
		log.Fatal("Please specify the databse host using environment variable DBHOST.")
	}
	if os.Getenv("DBNAME") == "" {
		log.Fatal("Please specify the name of the database using environment variable DBNAME.")
	}

	// Extracts the PORT, VIDEO_STORAGE_HOST and VIDEO_STORAGE_PORT environment variable.
	PORT := os.Getenv("PORT")
	VIDEO_STORAGE_HOST := os.Getenv("VIDEO_STORAGE_HOST")
	VIDEO_STORAGE_PORT := os.Getenv("VIDEO_STORAGE_PORT")
	DBHOST := os.Getenv("DBHOST")
	DBNAME := os.Getenv("DBNAME")

	log.Printf("Forwarding video requests to %s:%s.", VIDEO_STORAGE_HOST, VIDEO_STORAGE_PORT)

	// Create a MongoDb client connecting to the Database
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(DBHOST))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Fatalf("Failed to disconnect with MongoDB: %v", err)
		}
	}()

	// Get the database
	db := client.Database(DBNAME)
	// Get the video collections
	videosCollection := db.Collection("videos")

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World!")
	})

	// Registers a HTTP GET route for video streaming.
	r.GET("/video", func(c *gin.Context) {
		videoId := c.Query("id")
		if videoId == "" {
			c.JSON(http.StatusNotFound, "The video was not found")
			return
		}

		videoObjectId, err := primitive.ObjectIDFromHex(videoId)
		if err != nil {
			log.Println("Invalid id")
		}

		// Retrieves the first matching document
		var videoRecord VideoRecord
		err = videosCollection.FindOne(context.Background(), bson.M{"_id": videoObjectId}).Decode(&videoRecord)

		// Prints a message if no documents are matched or if any
		// other errors occur during the operation
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, "Video not found")
				return
			}
			c.JSON(http.StatusInternalServerError, "Error retrieving video")
			return
		}
		log.Printf("Translated id %v to path %v.", videoId, videoRecord.VideoPath)

		// Created a Director for reverse proxy
		director := func(req *http.Request) {
			req.URL.Host = VIDEO_STORAGE_HOST + ":" + VIDEO_STORAGE_PORT
			req.URL.Path = "/video"
			req.URL.RawQuery = "path=" + videoRecord.VideoPath
			req.URL.Scheme = "http"
			req.Host = VIDEO_STORAGE_HOST + ":" + VIDEO_STORAGE_PORT
		}

		// Created a reverse proxy
		proxy := &httputil.ReverseProxy{Director: director}
		proxy.ServeHTTP(c.Writer, c.Request)
	})

	// Starts the HTTP server.
	log.Println("Microservice listening, please load the data file db-fixture/videos.json into your database before testing this microservice.")
	err = r.Run(":" + PORT) // listen and serve on 0.0.0.0:3000

	if err != nil {
		log.Fatalf("Microservice failed to start: %v", err)
	}
}
