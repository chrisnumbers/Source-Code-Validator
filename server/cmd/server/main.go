package main

import (
	"context"
	"fmt"
	"github.com/dotenv-org/godotenvvault"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"source-code-validator/server/internal/api"
	"source-code-validator/server/internal/util"
	"strconv"
	"time"
)

func main() {
	// Load environment variables
	err := godotenvvault.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_URI"))
	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to MongoDB
	mongoClient, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("MongoDB connection error: %v", err)
	}

	// Ping the database to verify connection
	if err := mongoClient.Ping(ctx, nil); err != nil {
		log.Fatalf("MongoDB ping error: %v", err)
	}
	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			log.Fatal(err)
		}
		fmt.Println("🔌 Disconnected from MongoDB.")
	}()

	redisDB, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		log.Fatalf("Invalid REDIS_DB value: %v", err)
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),     // Redis server address
		Password: os.Getenv("REDIS_PASSWORD"), // No password set
		DB:       redisDB,                     // Use default MongoDB
	})

	// Test the connection with PING
	pong, err := redisClient.Ping(ctx).Result()
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to Redis: %v", err))
	}
	fmt.Println("Redis connected successfully:", pong)

	router := api.SetupRouter(util.NewHandler(mongoClient, redisClient))

	err = router.Run("0.0.0.0:8080")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
		return

	} // Run on port 8080

}
