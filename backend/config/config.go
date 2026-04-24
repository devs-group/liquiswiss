package config

import (
	"os"
)

type Config struct {
	WebHost            string
	DBUser             string
	DBPassword         string
	DBHost             string
	DBPort             string
	DBName             string
	JWTKey             []byte
	SendgridToken      string
	SendgridTemplateID string
	FixerIOURl         string
	FixerIOKey         string

	AigentAPIURL       string
	AigentClientID     string
	AigentClientSecret string

	BackendPublicURL string
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

		SendgridToken:      getEnv("SEND_GRID_TOKEN", ""),
		SendgridTemplateID: getEnv("SEND_GRID_TEMPLATE_ID", ""),

		FixerIOURl: getEnv("FIXER_IO_URL", ""),
		FixerIOKey: getEnv("FIXER_IO_KEY", ""),

		AigentAPIURL:       getEnv("AIGENT_API_URL", ""),
		AigentClientID:     getEnv("AIGENT_CLIENT_ID", ""),
		AigentClientSecret: getEnv("AIGENT_CLIENT_SECRET", ""),

		BackendPublicURL: getEnv("BACKEND_PUBLIC_URL", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
