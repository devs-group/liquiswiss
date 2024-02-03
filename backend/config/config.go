package config

import (
	"os"
)

type Config struct {
	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     string
	DBName     string
}

func GetConfig() Config {
	return Config{
		DBUser:     GetEnv("DB_USER", "ls_user"),
		DBPassword: GetEnv("DB_PASSWORD", "password"),
		DBHost:     GetEnv("DB_HOST", "localhost"),
		DBPort:     GetEnv("DB_PORT", "3306"),
		DBName:     GetEnv("DB_NAME", "liquiswiss"),
	}
}

func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
