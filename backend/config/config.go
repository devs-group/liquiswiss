package config

import (
	"os"
	"strconv"
)

type Config struct {
	WebHost         string
	DBUser          string
	DBPassword      string
	DBHost          string
	DBPort          string
	DBName          string
	JWTKey          []byte
	SMTPHost        string
	SMTPPort        int
	SMTPUser        string
	SMTPPass        string
	SMTPFromAddress string
	SMTPFromName    string
	SMTPTLS         string
	FixerIOURl      string
	FixerIOKey      string
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

		SMTPHost:        getEnv("SMTP_HOST", ""),
		SMTPPort:        getEnvInt("SMTP_PORT", 1025),
		SMTPUser:        getEnv("SMTP_USER", ""),
		SMTPPass:        getEnv("SMTP_PASS", ""),
		SMTPFromAddress: getEnv("SMTP_FROM_ADDRESS", "no-reply@liquiswiss.local"),
		SMTPFromName:    getEnv("SMTP_FROM_NAME", "LiquiSwiss"),
		SMTPTLS:         getEnv("SMTP_TLS", "off"),

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

func getEnvInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if n, err := strconv.Atoi(value); err == nil {
			return n
		}
	}
	return fallback
}
