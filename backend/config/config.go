package config

import (
	"os"
)

type Config struct {
	WebHost       string
	DBUser        string
	DBPassword    string
	DBHost        string
	DBPort        string
	DBName        string
	JWTKey        []byte
	SendgridToken string
	FixerIOURl    string
	FixerIOKey    string
}

func GetConfig() Config {
	return Config{
		WebHost: getEnv("WEB_HOST", "http://localhost:3000"),

		DBUser:     getEnv("DB_USER", ""),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBHost:     getEnv("DB_HOST", ""),
		DBPort:     getEnv("DB_PORT", ""),
		DBName:     getEnv("DB_NAME", ""),

		JWTKey: []byte(getEnv("JWT_KEY", "")),

		SendgridToken: getEnv("SEND_GRID_TOKEN", ""),

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
