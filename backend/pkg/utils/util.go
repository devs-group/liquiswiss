package utils

import (
	"encoding/json"
	"github.com/google/uuid"
	"os"
	"reflect"
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

func IsStructEmpty(v interface{}) bool {
	val := reflect.ValueOf(v)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return false
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if !field.IsNil() {
			return false
		}
	}
	return true
}

func StringAsPointer(s string) *string {
	return &s
}
