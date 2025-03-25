package utils

import (
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"
)

const (
	envPath = "../../.env"
)

func InitEnv() {
	if err := godotenv.Load(envPath); err != nil {
		log.Print("No .env file found")
	}
}

func GetEnvWrapper(key string) (string, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return "", errors.New("Failed to get TOKEN env variable.")
	}
	return value, nil
}
