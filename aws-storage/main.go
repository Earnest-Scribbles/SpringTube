package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

// Create the S3 Client to communicate with AWS S3
func createS3Service() *s3.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Println("error:", err)
		return nil
	}

	s3Client := s3.NewFromConfig(cfg)
	return s3Client
}

func main() {
	// Throws an error if the any required environment variables are missing.
	if os.Getenv("PORT") == "" {
		log.Fatal("Please specify the port number for the HTTP server with the environment variable PORT.")
	}

	if os.Getenv("AWS_REGION") == "" {
		log.Fatal("Please specify the AWS Region with the environment variable AWS_REGION.")
	}

	if os.Getenv("AWS_ACCESS_KEY_ID") == "" {
		log.Fatal("Please specify the access key ID to AWS Account with the environment variable AWS_ACCESS_KEY_ID.")
	}

	if os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
		log.Fatal("Please specify the Secret access key to AWS Account with the environment variable AWS_SECRET_ACCESS_KEY.")
	}

	if os.Getenv("AWS_SESSION_TOKEN") == "" {
		log.Fatal("Please specify the Session token to AWS Account with the environment variable AWS_SESSION_TOKEN.")
	}

	// Extracts the environment variables.
	PORT := os.Getenv("PORT")

	log.Println("Serving videos from AWS S3")

	r := gin.Default()

	// Registers a HTTP GET route to retrieve videos from storage.
	r.GET("/video", func(c *gin.Context) {
		// Path query parameter is the video path
		videoPath := c.Query("path")
		if videoPath == "" {
			c.JSON(http.StatusBadRequest, "Path parameter is required")
			return
		}
		log.Printf("Serving videos from path %s.", videoPath)

		s3Client := createS3Service()

		bucketName := "springtube"
		objectKey := "videos/" + videoPath

		s3Object, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(objectKey),
		})

		if err != nil {
			log.Printf("Couldn't get object from %v:%v. Here's why: %v\n",
				bucketName, objectKey, err)
		}

		defer s3Object.Body.Close()

		// read entire body from s3 object
		body, err := io.ReadAll(s3Object.Body)
		if err != nil {
			log.Printf("Couldn't read object body from %v. Here's why: %v\n", objectKey, err)
		}

		// c.Header("Content-Length", strconv.Itoa(int(*s3Object.ContentLength)))
		// c.Header("Content-Type", "video/mp4")

		// // Add the code to stream the file into response
		// c.Stream(func(w io.Writer) bool {
		// 	_, err := io.Copy(w, s3Object.Body)
		// 	if err != nil {
		// 		log.Printf("Failed to stream video: %v", err)
		// 		return false
		// 	}
		// 	return false // Returning false after streaming is done
		// })

		// Add the video creating a new reader from the body to response stream adding in content length and type headers
		c.DataFromReader(http.StatusOK, *s3Object.ContentLength, *s3Object.ContentType, bytes.NewReader(body), nil)
	})

	// Starts the HTTP server.
	log.Println("Microservice online")
	err := r.Run(":" + PORT) // listen and serve on 0.0.0.0:3000

	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
