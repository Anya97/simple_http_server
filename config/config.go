package config

import (
	"log"
	"os"
)

type Config struct {
	Port string
}

func GetConfig() *Config {
	return &Config{
		Port: getEnv("PORT", "3333"),
	}
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	log.Print("No .env file found")
	return defaultVal
}
