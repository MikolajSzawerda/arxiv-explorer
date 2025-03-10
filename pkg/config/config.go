package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	OpenAIKey string
}

func LoadConfig() (*Config, error) {
	// Load .env file if present
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, continuing without it.")
	}

	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not set")
	}

	return &Config{
		OpenAIKey: key,
	}, nil
}
