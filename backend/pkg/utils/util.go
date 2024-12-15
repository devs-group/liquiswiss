package utils

import (
	"encoding/json"
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

func StructToMap[T any](r T) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	var templateData map[string]interface{}
	err = json.Unmarshal(jsonData, &templateData)
	if err != nil {
		return nil, err
	}

	return templateData, nil
}
