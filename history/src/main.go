package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Structure to hold the video path from video streaming microservice
type VideoHistory struct {
	VideoPath string `json:"videoPath"`
}

func main() {
	// Throws an error if the any required environment variables are missing.
	if os.Getenv("PORT") == "" {
		log.Fatal("Please specify the port number for the HTTP server with the environment variable PORT.")
	}
	if os.Getenv("DBHOST") == "" {
		log.Fatal("Please specify the database host using environment variable DBHOST.")
	}
	if os.Getenv("DBNAME") == "" {
		log.Fatal("Please specify the name of the database using environment variable DBNAME.")
	}

	// Extracts the environment variables.
	PORT := os.Getenv("PORT")
	DBHOST := os.Getenv("DBHOST")
	DBNAME := os.Getenv("DBNAME")

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
	// Gets the collection for storing video viewing history.
	historyCollection := db.Collection("history")

	r := gin.Default()

	// Handles HTTP POST request to /viewed.
	r.POST("/viewed", func(c *gin.Context) {
		var videoHistory VideoHistory

		// Read JSON body from HTTP request.
		err := c.BindJSON(&videoHistory)
		if err != nil {
			log.Printf("Invalid JSON body: %v", err)
			c.Status(http.StatusBadRequest)
			return
		}

		_, err = historyCollection.InsertOne(context.Background(), bson.M{"videoPath": videoHistory.VideoPath})
		if err != nil {
			log.Printf("Error inserting document: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}

		log.Printf("Added video %s to history.", videoHistory.VideoPath)
		c.Status(http.StatusOK)
	})

	// Starts the HTTP server.
	log.Println("Microservice online")
	err = r.Run(":" + PORT)

	if err != nil {
		log.Fatalf("Microservice failed to start: %v", err)
	}

}
