package utils

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

const (
	envPath = "../.env"
)

func CreateApp() *App {
	return &App{Config: *createConfig()}
}

func createConfig() *Config {
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

	stringURL, err := GetEnvWrapper("MM_URL")
	if err != nil {
		log.Fatal(err)
	}
	mmURL, err := url.Parse(stringURL)
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
		return "", fmt.Errorf("failed to get %s env variable", key)
	}
	return value, nil
}
