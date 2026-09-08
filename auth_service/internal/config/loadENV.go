package config

import (
	"log"

	"github.com/joho/godotenv"
)

func LoadEnv() bool {
	_ = godotenv.Load()

	return true
}

func IsEnvSet(key string, val string) {
	if val == "" {
		log.Fatalf("Key %s is not set inside .env", key)
	}
}
