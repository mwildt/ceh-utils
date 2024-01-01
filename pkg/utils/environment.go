package utils

import (
	"log"
	"os"
)

func GetEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func GetEnvOrFail(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	log.Fatalf("unable to find env value with key %s, exiting", key)
	return ""
}
