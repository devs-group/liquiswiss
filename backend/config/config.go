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
	JWTKey     []byte
	FixerIOURl string
	FixerIOKey string
}

func GetConfig() Config {
	return Config{
		DBUser:     getEnv("DB_USER", ""),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBHost:     getEnv("DB_HOST", ""),
		DBPort:     getEnv("DB_PORT", ""),
		DBName:     getEnv("DB_NAME", ""),

		JWTKey: []byte(getEnv("JWT_KEY", "")),

		FixerIOURl: getEnv("FIXER_IO_URL", ""),
		FixerIOKey: getEnv("FIXER_IO_KEY", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
