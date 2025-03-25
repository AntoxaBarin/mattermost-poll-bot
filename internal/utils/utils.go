package utils

import (
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"
)

const (
	envPath = "../.env"
)

func CreateConfig() *Config {
	if err := godotenv.Load(envPath); err != nil {
		log.Print("No .env file found")
	}

	botToken, err := GetEnvWrapper("BOT_TOKEN")
	if err != nil {
		log.Fatal(err)
	}

	botPort, err := GetEnvWrapper("BOT_PORT")
	if err != nil {
		log.Fatal(err)
	}

	mmURL, err := GetEnvWrapper("MM_URL")
	if err != nil {
		log.Fatal(err)
	}

	return &Config{
		URL:      mmURL,
		BotToken: botToken,
		BotPort:  botPort,
	}
}

func GetEnvWrapper(key string) (string, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return "", errors.New("Failed to get TOKEN env variable.")
	}
	return value, nil
}
