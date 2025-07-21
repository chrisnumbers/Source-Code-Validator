package main

import (
	"github.com/dotenv-org/godotenvvault"
	"log"
	"source-code-validator/server/internal/api"
)

func main() {
	// Load environment variables
	err := godotenvvault.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		return
	}

	router := api.SetupRouter()
	err = router.Run(":8080")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
		return
	} // Run on port 8080
}
