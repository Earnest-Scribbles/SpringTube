package main

import (
	"bytes"
	"context"
	"encoding/json"
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

// Structure to hold the video id and path from mongodb
type VideoRecord struct {
	ID        string `bson:"_id"`
	VideoPath string `bson:"videoPath"`
}

// Send the "viewed" to the history microservice.
func sendViewedMessage(videoPath string) {
	requestBody, err := json.Marshal(map[string]string{
		"videoPath": videoPath,
	})
	if err != nil {
		log.Fatalf("Failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest("POST", "http://history/viewed", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Println("Error creating request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to send 'viewed' message: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		log.Println("Sent 'viewed' message to history microservice.")
	} else {
		log.Println("Failed to send 'viewed' message!")
	}
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
		log.Fatal("Please specify the database host using environment variable DBHOST.")
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

	// Gets the database for this microservice
	db := client.Database(DBNAME)
	// Gets the collection for storing video path.
	videosCollection := db.Collection("videos")

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World!")
	})

	// Registers a HTTP GET route for video streaming.
	r.GET("/video", func(c *gin.Context) {
		// convert the id passed into ObjectID which mongoDB can understand
		videoId, err := primitive.ObjectIDFromHex(c.Query("id"))
		if err != nil {
			c.JSON(http.StatusNotFound, "The video was not found")
			log.Printf("Invalid id: %v", err)
			return
		}

		// Retrieves the first matching document
		var videoRecord VideoRecord
		err = videosCollection.FindOne(context.Background(), bson.M{"_id": videoId}).Decode(&videoRecord)

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

		// Sends the "viewed" message to indicate this video has been watched.
		sendViewedMessage("./videos/" + videoRecord.VideoPath)
	})

	// Starts the HTTP server.
	log.Println("Microservice listening, please load the data file db-fixture/videos.json into your database before testing this microservice.")
	err = r.Run(":" + PORT) // listen and serve on 0.0.0.0:3000

	if err != nil {
		log.Fatalf("Microservice failed to start: %v", err)
	}
}
