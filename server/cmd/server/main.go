package main

import (
	"context"
	"fmt"
	"github.com/dotenv-org/godotenvvault"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"source-code-validator/server/internal/api"
	"source-code-validator/server/internal/util"
	"time"
)

func main() {
	// Load environment variables
	err := godotenvvault.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		return
	}

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("MongoDB connection error: %v", err)
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("MongoDB ping error: %v", err)
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Fatal(err)
		}
		fmt.Println("🔌 Disconnected from MongoDB.")
	}()

	router := api.SetupRouter(util.NewHandler(client))

	err = router.Run(":8080")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
		return

	} // Run on port 8080
}
