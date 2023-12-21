package config

import (
	"fmt"
	"log"
	"os"
	"strings"
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
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		fmt.Println(pair[0])
	}

	log.Print("No .env file found")
	return defaultVal
}
