package main

import (
	"context"
	"math/rand"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func initData(redisClient *redis.Client) {
    // Define gifs
    gifs := map[string]string{
        "1": "https://media.giphy.com/media/3oKIPEIA9vi2QFjHrq/giphy.gif",
        "2": "https://media.giphy.com/media/11sBLVxNs7v6WA/giphy.gif",
        "3": "https://media.giphy.com/media/l0MYt5jPR6QX5pnqM/giphy.gif",
        "4": "https://media.giphy.com/media/xUPGcgtKxmFzOjV7EI/giphy.gif",
        "5": "https://media.giphy.com/media/l1KVcAP6jyiH3z6U4/giphy.gif",
    }
    // Add gifs to Redis
    for gifID, gifURL := range gifs {
        err := redisClient.HSet(context.Background(), "gifs_tags", gifID, gifURL).Err()
        if err != nil {
            panic(err)
        }
        // Set the initial like count to 0
        err = redisClient.HSet(context.Background(), "gifs_likes", gifID, 0).Err()
        if err != nil {
            panic(err)
        }
        // Set the initial dislike count to 0
        err = redisClient.HSet(context.Background(), "gifs_dislikes", gifID, 0).Err()
        if err != nil {
            panic(err)
        }
    }
}

func main() {
	// Connect to the Redis database
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
        Password: "",
        DB:       0,
	})

	// Initialize the test data
    initData(redisClient)

	// Initialize the Gin router
	router := gin.Default()

	// Define endpoint for giving a like to a gif
	router.POST("/gifs/:url/like", func(c *gin.Context) {
		// Get the gif URL from the URL parameters
		gifURL := c.Param("url")
		// Increment the like count for the gif in Redis
		redisClient.HIncrBy(c, "gifs_likes", gifURL, 1)
	})

	// Define endpoint for giving a dislike to a gif
	router.POST("/gifs/:url/dislike", func(c *gin.Context) {
		// Get the gif URL from the URL parameters
		gifURL := c.Param("url")
		// Increment the dislike count for the gif in Redis
		redisClient.HIncrBy(c, "gifs_dislikes", gifURL, 1)
	})

	// Define endpoint for getting gifs
	router.GET("/gifs", func(c *gin.Context) {
        // Get the list of requested tags from the query parameters
        requestedTags := c.Query("tags")
        // Get all gif URLs and their tags from Redis
        gifTags, err := redisClient.HGetAll(c, "gifs_tags").Result()
        if err != nil {
            c.AbortWithStatus(http.StatusInternalServerError)
            return
        }
        // Filter the gif URLs based on the requested tags
        var filteredGifURLs []string
        for gifURL, gifTag := range gifTags {
            if gifTag == requestedTags {
                filteredGifURLs = append(filteredGifURLs, gifURL)
            }
        }
        // Select a random gif URL from the filtered list based on its popularity
        if len(filteredGifURLs) > 0 {
            gifURL := filteredGifURLs[rand.Intn(len(filteredGifURLs))]
            // Serve the gif URL as the HTTP response
            c.String(http.StatusOK, gifURL)
        } else {
            c.AbortWithStatus(http.StatusNotFound)
        }
    })

	// Start the HTTP server
	router.Run(":8080")
}