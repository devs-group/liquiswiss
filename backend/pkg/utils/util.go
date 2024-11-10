package utils

import (
	"github.com/google/uuid"
	"os"
)

func IsProduction() bool {
	mode, _ := os.LookupEnv("GIN_MODE")
	return mode == "release"
}

// GenerateUUID generates a new UUID
func GenerateUUID() string {
	return uuid.New().String()
}
